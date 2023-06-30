package set

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
	query, err := command.getOption(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapSetRequest(query, lg)
	err = command.requestManager.Request(s, i, setRequestRoutingKey, msg, command.getSetReply)
	if err != nil {
		panic(err)
	}
}

func (command *Command) updateSet(s *discordgo.Session, i *discordgo.InteractionCreate,
	lg discordgo.Locale) {
	// TODO
	data := i.MessageComponentData()
	fmt.Printf("coucou %v", data)
}

func (command *Command) getOption(ctx context.Context) (string, error) {
	query, ok := ctx.Value(constants.ContextKeyQuery).(string)
	if !ok {
		return "", fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyQuery))
	}

	return query, nil
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

func isAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_SET_ANSWER &&
		message.EncyclopediaSetAnswer != nil
}

func matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == contract.SetCommandName
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	// TODO
	return true
}
