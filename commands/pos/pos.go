package pos

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/checks"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(guildService guilds.Service, portalService portals.Service,
	serverService servers.Service, emojiService emojis.Service,
	requestManager requests.RequestManager) *Command {
	cmd := Command{
		AbstractCommand: commands.AbstractCommand{
			DiscordID: viper.GetString(constants.PosID),
		},
		guildService:   guildService,
		portalService:  portalService,
		serverService:  serverService,
		emojiService:   emojiService,
		requestManager: requestManager,
	}

	checkServer := checks.CheckServerWithFallback(contract.PosServerOptionName,
		cmd.serverService, cmd.guildService)

	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.Use(cmd.checkDimension,
			checkServer, cmd.request),
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
	}

	return &cmd
}

func (command *Command) GetName() string {
	return contract.PosCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        fmt.Sprintf("/%v", contract.PosCommandName),
			CommandID:   fmt.Sprintf("</%v:%v>", contract.PosCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed", contract.PosCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial", contract.PosCommandName)),
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
	dimension, server, err := getOptions(ctx)
	if err != nil {
		panic(err)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapPortalPositionRequest(dimension, server, authorID, i.Locale)
	err = command.requestManager.Request(s, i, constants.PortalRequestRoutingKey,
		msg, command.respond)
	if err != nil {
		panic(err)
	}
}

func (command *Command) respond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	embeds := make([]*discordgo.MessageEmbed, 0)
	for _, position := range message.GetPortalPositionAnswer().GetPositions() {
		embeds = append(embeds, mappers.MapPortalToEmbed(position, command.portalService,
			command.serverService, command.emojiService, message.Language))
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds})
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func getOptions(ctx context.Context) (
	entities.Dimension, entities.Server, error) {
	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.Dimension{}, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	dimension := entities.Dimension{}
	if ctx.Value(constants.ContextKeyDimension) != nil {
		dimension, ok = ctx.Value(constants.ContextKeyDimension).(entities.Dimension)
		if !ok {
			return entities.Dimension{}, entities.Server{},
				fmt.Errorf("cannot cast %v as entities.Dimension", ctx.Value(constants.ContextKeyDimension))
		}
	}

	return dimension, server, nil
}

func isAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_PORTAL_POSITION_ANSWER &&
		message.PortalPositionAnswer != nil
}
