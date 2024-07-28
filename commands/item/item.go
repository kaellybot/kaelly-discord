package item

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(characService characteristics.Service, emojiService emojis.Service,
	requestManager requests.RequestManager) *Command {
	cmd := Command{
		characService:  characService,
		emojiService:   emojiService,
		requestManager: requestManager,
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
			CommandID:   "</item:1116290248587100251>",
			Description: i18n.Get(lg, "item.help.detailed"),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return command.matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command.CallHandler(s, i, command.handlers)
}

func (command *Command) getItem(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	query, err := getQueryOption(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapItemRequest(query, false, amqp.ItemType_EQUIPMENT, i.Locale)
	err = command.requestManager.Request(s, i, itemRequestRoutingKey, msg, command.getItemReply)
	if err != nil {
		panic(err)
	}
}

func (command *Command) updateItem(s *discordgo.Session,
	i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	properties := make(map[string]any)
	var query, itemType string
	var ok bool
	if itemType, ok = contract.ExtractItemCustomID(customID); ok {
		values := i.MessageComponentData().Values
		if len(values) != 1 {
			log.Error().
				Str(constants.LogCommand, command.GetName()).
				Str(constants.LogCustomID, customID).
				Msgf("Cannot retrieve item ID from value, panicking...")
			panic(commands.ErrInvalidInteraction)
		}
		query = values[0]
		properties[isRecipeProperty] = false
	} else if query, itemType, ok = contract.ExtractItemEffectsCustomID(customID); ok {
		properties[isRecipeProperty] = false
	} else if query, itemType, ok = contract.ExtractItemRecipeCustomID(customID); ok {
		properties[isRecipeProperty] = true
	} else {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	itemTypeID, found := amqp.ItemType_value[itemType]
	if !found {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot retrieve item type from custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	msg := mappers.MapItemRequest(query, true, amqp.ItemType(itemTypeID), i.Locale)
	err := command.requestManager.Request(s, i, itemRequestRoutingKey,
		msg, command.updateItemReply, properties)
	if err != nil {
		panic(err)
	}
}

func (command *Command) getItemReply(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	command.updateItemReply(ctx, s, i, message, map[string]any{isRecipeProperty: false})
}

func (command *Command) updateItemReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {
	if !isAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	isRecipeValue, found := properties[isRecipeProperty]
	if !found {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogRequestProperty, isRecipeProperty).
			Msgf("Cannot find request property, panicking...")
		panic(commands.ErrRequestPropertyNotFound)
	}
	isRecipe, ok := isRecipeValue.(bool)
	if !ok {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogRequestProperty, isRecipeProperty).
			Msgf("Cannot convert request property, panicking...")
		panic(commands.ErrRequestPropertyNotFound)
	}

	reply := mappers.MapItemToWebhookEdit(message.GetEncyclopediaItemAnswer(), isRecipe,
		command.characService, command.emojiService, message.Language)

	_, err := s.InteractionResponseEdit(i.Interaction, reply)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
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
