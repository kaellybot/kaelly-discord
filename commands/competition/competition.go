package competition

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(emojiService emojis.Service, requestManager requests.RequestManager,
) *Command {
	cmd := Command{
		requestManager: requestManager,
		emojiService:   emojiService,
	}

	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.Use(
			cmd.checkOptionalMapNumber, cmd.request),
	}

	return &cmd
}

func (command *Command) GetName() string {
	return contract.MapCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        "/map",
			CommandID:   "</map:1291722831767404667>",
			Description: i18n.Get(lg, "map.help.detailed"),
			TutorialURL: i18n.Get(lg, "map.help.tutorial"),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command.CallHandler(s, i, command.handlers)
}

func (command *Command) request(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	mapNumber, err := getOptions(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapCompetitionMapRequest(mapNumber, i.Locale)
	err = command.requestManager.Request(s, i, competitionRequestRoutingKey, msg, command.respond)
	if err != nil {
		panic(err)
	}
}

func (command *Command) respond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	_, err := s.InteractionResponseEdit(i.Interaction,
		mappers.MapCompetitionMapToWebhookEdit(message.GetCompetitionMapAnswer(),
			constants.MapTypeNormal, command.emojiService, message.Language))
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func getOptions(ctx context.Context) (int64, error) {
	mapNumber, ok := ctx.Value(constants.ContextKeyMap).(int64)
	if !ok {
		return -1,
			fmt.Errorf("cannot cast %v as int64", ctx.Value(constants.ContextKeyMap))
	}

	return mapNumber, nil
}

func isAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_COMPETITION_MAP_ANSWER &&
		message.CompetitionMapAnswer != nil
}
