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

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	command.CallHandler(s, i, lg, command.handlers)
}

func (command *Command) getItem(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, _ middlewares.NextFunc) {
	query, err := getQueryOption(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapItemRequest(query, false, lg)
	err = command.requestManager.Request(s, i, itemRequestRoutingKey, msg, command.getItemReply)
	if err != nil {
		panic(err)
	}
}

func (command *Command) updateItem(s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {
	customID := i.MessageComponentData().CustomID
	properties := make(map[string]any)
	var query string
	if contract.ExtractItemCustomID(customID) {
		values := i.MessageComponentData().Values
		if len(values) != 1 {
			log.Error().
				Str(constants.LogCommand, contract.ItemCommandName).
				Str(constants.LogCustomID, customID).
				Msgf("Cannot retrieve item ID from value, panicking...")
			panic(commands.ErrInvalidInteraction)
		}
		query = values[0]
		properties[isRecipeProperty] = false
	} else if itemID, ok := contract.ExtractItemEffectsCustomID(customID); ok {
		query = itemID
		properties[isRecipeProperty] = false
	} else if itemID, ok = contract.ExtractItemRecipeCustomID(customID); ok {
		query = itemID
		properties[isRecipeProperty] = true
	} else {
		log.Error().
			Str(constants.LogCommand, contract.ItemCommandName).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	msg := mappers.MapItemRequest(query, true, lg)
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
			Str(constants.LogCommand, contract.ItemCommandName).
			Str(constants.LogRequestProperty, isRecipeProperty).
			Msgf("Cannot find request property, panicking...")
		panic(commands.ErrRequestPropertyNotFound)
	}
	isRecipe, ok := isRecipeValue.(bool)
	if !ok {
		log.Error().
			Str(constants.LogCommand, contract.ItemCommandName).
			Str(constants.LogRequestProperty, isRecipeProperty).
			Msgf("Cannot convert request property, panicking...")
		panic(commands.ErrRequestPropertyNotFound)
	}

	reply := mappers.MapItemToWebhookEdit(message.EncyclopediaItemAnswer, isRecipe,
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
		message.EncyclopediaItemAnswer != nil
}

func getQueryOption(ctx context.Context) (string, error) {
	query, ok := ctx.Value(constants.ContextKeyQuery).(string)
	if !ok {
		return "", fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyQuery))
	}

	return query, nil
}

func matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == contract.ItemCommandName
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToItem(i.MessageComponentData().CustomID)
}
