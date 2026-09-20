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
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	di18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//nolint:exhaustive // only the useful interaction types are handled.
func New(characService characteristics.Service, emojiService emojis.Service,
	requestManager requests.RequestManager) *Command {
	cmd := Command{
		DiscordID:      viper.GetString(constants.SetID),
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

func (command *Command) GetName() string {
	return contract.SetCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        fmt.Sprintf("/%v", contract.SetCommandName),
			CommandID:   fmt.Sprintf("</%v:%v>", contract.SetCommandName, command.DiscordID),
			Description: di18n.Get(lg, fmt.Sprintf("%v.help.detailed", contract.SetCommandName)),
			TutorialURL: di18n.Get(lg, fmt.Sprintf("%v.help.tutorial", contract.SetCommandName)),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return command.matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return command.CallHandler(s, i, command.handlers)
}

func (command *Command) getSet(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	query, err := getQueryOption(ctx)
	if err != nil {
		return err
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapItemRequest(query, false, amqp.ItemType_SET_TYPE, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.SetRequestRoutingKey,
		msg, command.getSetReply)
}

func (command *Command) updateSet(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	properties := make(map[string]any)
	var query string
	var callback requests.RequestCallback
	if setID, ok := contract.ExtractSetCustomID(customID); ok {
		query = setID
		callback = command.getSetReply
	} else if setID, ok = contract.ExtractSetBonusCustomID(customID); ok {
		query = setID
		callback = command.updateSetReply
		itemNumber, err := getBonusValue(i.MessageComponentData())
		if err != nil {
			return fmt.Errorf("%w: custom ID %q", err, customID)
		}

		properties[itemNumberProperty] = itemNumber
	} else {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapItemRequest(query, true, amqp.ItemType_SET_TYPE, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.SetRequestRoutingKey,
		msg, callback, properties)
}

func (command *Command) getSetReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) error {
	if !isAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	if message.GetEncyclopediaItemAnswer().GetSet() == nil {
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

	reply := mappers.MapSetToDefaultWebhookEdit(message.EncyclopediaItemAnswer,
		command.characService, command.emojiService, message.Language)
	_, err := s.InteractionResponseEdit(i.Interaction, reply)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}

func (command *Command) updateSetReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) error {
	if !isAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	itemNumber, ok := properties[itemNumberProperty].(int)
	if !ok {
		return fmt.Errorf("%w: %v", commands.ErrRequestPropertyNotFound, itemNumberProperty)
	}

	reply := mappers.MapSetToWebhookEdit(message.EncyclopediaItemAnswer, itemNumber,
		command.characService, command.emojiService, message.Language)
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
		message.EncyclopediaItemAnswer != nil &&
		message.EncyclopediaItemAnswer.GetType() == amqp.ItemType_SET_TYPE
}

func getQueryOption(ctx context.Context) (string, error) {
	query, ok := ctx.Value(constants.ContextKeyQuery).(string)
	if !ok {
		return "", fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyQuery))
	}

	return query, nil
}

// getBonusValue reads the number of items the caller selected. Anything else
// than one parsable number means the component is not one we can read.
func getBonusValue(data discordgo.MessageComponentInteractionData) (int, error) {
	values := data.Values
	if len(values) != 1 {
		return 0, fmt.Errorf("%w: selected %v values instead of one",
			commands.ErrInvalidInteraction, len(values))
	}

	itemNumber, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, fmt.Errorf("%w: %w", commands.ErrInvalidInteraction, err)
	}

	return itemNumber, nil
}

func (command *Command) matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToSet(i.MessageComponentData().CustomID)
}
