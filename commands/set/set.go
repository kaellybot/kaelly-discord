package set

import (
	"context"
	"fmt"
	"strconv"

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
		discordgo.InteractionApplicationCommand:             middlewares.Use(cmd.checkQuery, cmd.getSet),
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
		discordgo.InteractionMessageComponent:               cmd.updateSet,
	}

	return &cmd
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	command.CallHandler(s, i, lg, command.handlers)
}

func (command *Command) getSet(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, _ middlewares.NextFunc) {
	query, err := getOption(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapSetRequest(query, false, lg)
	err = command.requestManager.Request(s, i, setRequestRoutingKey, msg, command.getSetReply)
	if err != nil {
		panic(err)
	}
}

func (command *Command) updateSet(s *discordgo.Session, i *discordgo.InteractionCreate,
	lg discordgo.Locale) {
	customID := i.MessageComponentData().CustomID
	properties := make(map[string]any)
	var query string
	var isID bool
	var callback requests.RequestCallback
	if setID, ok := contract.ExtractSetCustomID(customID); ok {
		query = setID
		callback = command.getSetReply
	} else if setID, ok = contract.ExtractSetBonusCustomID(customID); ok {
		query = setID
		isID = true
		callback = command.updateSetReply
		itemNumber, err := getBonusValue(i.MessageComponentData())
		if err != nil {
			log.Error().
				Str(constants.LogCommand, contract.SetCommandName).
				Str(constants.LogCustomID, customID).
				Str(constants.LogRequestProperty, itemNumberProperty).
				Strs(constants.LogRequestValue, i.MessageComponentData().Values).
				Msgf("Cannot retrieve itemNumber from values selected by user, panicking...")
			panic(err)
		}

		properties[itemNumberProperty] = itemNumber
	} else {
		log.Error().
			Str(constants.LogCommand, contract.SetCommandName).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	msg := mappers.MapSetRequest(query, isID, lg)
	err := command.requestManager.Request(s, i, setRequestRoutingKey,
		msg, callback, properties)
	if err != nil {
		panic(err)
	}
}

func (command *Command) getSetReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	reply := mappers.MapSetToDefaultWebhookEdit(message.EncyclopediaSetAnswer,
		command.characService, command.emojiService, message.Language)
	_, err := s.InteractionResponseEdit(i.Interaction, reply)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func (command *Command) updateSetReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {
	if !isAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	itemNumberValue, found := properties[itemNumberProperty]
	if !found {
		log.Error().
			Str(constants.LogCommand, contract.SetCommandName).
			Str(constants.LogRequestProperty, itemNumberProperty).
			Msgf("Cannot find request property, panicking...")
		panic(commands.ErrRequestPropertyNotFound)
	}
	itemNumber, ok := itemNumberValue.(int)
	if !ok {
		log.Error().
			Str(constants.LogCommand, contract.SetCommandName).
			Str(constants.LogRequestProperty, itemNumberProperty).
			Msgf("Cannot convert request property, panicking...")
		panic(commands.ErrRequestPropertyNotFound)
	}

	reply := mappers.MapSetToWebhookEdit(message.EncyclopediaSetAnswer, itemNumber,
		command.characService, command.emojiService, message.Language)
	_, err := s.InteractionResponseEdit(i.Interaction, reply)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func isAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_SET_ANSWER &&
		message.EncyclopediaSetAnswer != nil
}

func getOption(ctx context.Context) (string, error) {
	query, ok := ctx.Value(constants.ContextKeyQuery).(string)
	if !ok {
		return "", fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyQuery))
	}

	return query, nil
}

func getBonusValue(data discordgo.MessageComponentInteractionData) (int, error) {
	values := data.Values
	if len(values) != 1 {
		return 0, commands.ErrInvalidInteraction
	}
	return strconv.Atoi(values[0])
}

func matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == contract.SetCommandName
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToSet(i.MessageComponentData().CustomID)
}
