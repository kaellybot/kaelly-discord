package item

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/equipments"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	di18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//nolint:exhaustive // only the useful interaction types are handled.
func New(characService characteristics.Service, equipmentService equipments.Service,
	emojiService emojis.Service, requestManager requests.RequestManager) *Command {
	cmd := Command{
		DiscordID:        viper.GetString(constants.ItemID),
		characService:    characService,
		equipmentService: equipmentService,
		emojiService:     emojiService,
		requestManager:   requestManager,
	}

	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand:             middlewares.Use(cmd.checkQuery, cmd.getItem),
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
		discordgo.InteractionMessageComponent:               cmd.updateItem,
	}

	return &cmd
}

func (command *Command) GetName() string {
	return contract.ItemCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        fmt.Sprintf("/%v", contract.ItemCommandName),
			CommandID:   fmt.Sprintf("</%v:%v>", contract.ItemCommandName, command.DiscordID),
			Description: di18n.Get(lg, fmt.Sprintf("%v.help.detailed", contract.ItemCommandName)),
			TutorialURL: di18n.Get(lg, fmt.Sprintf("%v.help.tutorial", contract.ItemCommandName)),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return command.matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return command.CallHandler(s, i, command.handlers)
}

func (command *Command) getItem(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	query, err := getQueryOption(ctx)
	if err != nil {
		return err
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapItemRequest(query, false, amqp.ItemType_ANY_ITEM_TYPE, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.ItemRequestRoutingKey,
		msg, command.getItemReply)
}

func (command *Command) updateItem(s *discordgo.Session,
	i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	properties := make(map[string]any)
	var query, itemType string
	var ok bool
	if itemType, ok = contract.ExtractItemCustomID(customID); ok {
		values := i.MessageComponentData().Values
		if len(values) != 1 {
			return fmt.Errorf("%w: custom ID %q selected %v values instead of one",
				commands.ErrInvalidInteraction, customID, len(values))
		}
		query = values[0]
		properties[isRecipeProperty] = false
	} else if query, itemType, ok = contract.ExtractItemEffectsCustomID(customID); ok {
		properties[isRecipeProperty] = false
	} else if query, itemType, ok = contract.ExtractItemRecipeCustomID(customID); ok {
		properties[isRecipeProperty] = true
	} else {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	itemTypeID, found := amqp.ItemType_value[itemType]
	if !found {
		return fmt.Errorf("%w: custom ID %q carries unknown item type %q",
			commands.ErrInvalidInteraction, customID, itemType)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapItemRequest(query, true, amqp.ItemType(itemTypeID), authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.ItemRequestRoutingKey,
		msg, command.updateItemReply, properties)
}

func (command *Command) getItemReply(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) error {
	return command.updateItemReply(ctx, s, i, message, map[string]any{isRecipeProperty: false})
}

func (command *Command) updateItemReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) error {
	if !isAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	if message.GetEncyclopediaItemAnswer().GetEquipment() == nil {
		query := message.GetEncyclopediaItemAnswer().Query
		lg := i18n.MapAMQPLocale(message.Language)
		reply := mappers.MapQueryMismatch(query, lg)
		_, err := s.InteractionResponseEdit(i.Interaction, reply)
		if err != nil {
			log.Warn().Err(err).
				Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}

		return nil
	}

	isRecipe, ok := properties[isRecipeProperty].(bool)
	if !ok {
		return fmt.Errorf("%w: %v", commands.ErrRequestPropertyNotFound, isRecipeProperty)
	}

	reply := mappers.MapItemToWebhookEdit(message.GetEncyclopediaItemAnswer(), isRecipe,
		command.characService, command.equipmentService, command.emojiService, message.Language)

	_, err := s.InteractionResponseEdit(i.Interaction, reply)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}

func isAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER &&
		message.GetEncyclopediaItemAnswer() != nil
}

func getQueryOption(ctx context.Context) (string, error) {
	query, ok := ctx.Value(constants.ContextKeyQuery).(string)
	if !ok {
		return "", fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyQuery))
	}

	return query, nil
}

func (command *Command) matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToItem(i.MessageComponentData().CustomID)
}
