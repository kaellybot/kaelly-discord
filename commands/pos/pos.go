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
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/rs/zerolog/log"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(guildService guilds.Service, portalService portals.Service,
	serverService servers.Service, requestManager requests.RequestManager) *Command {
	command := Command{
		guildService:   guildService,
		portalService:  portalService,
		serverService:  serverService,
		requestManager: requestManager,
	}

	command.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.Use(command.checkDimension,
			command.checkServer, command.request),
		discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
	}

	return &command
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return i.ApplicationCommandData().Name == contract.PosCommandName
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	command.CallHandler(s, i, lg, command.handlers)
}

func (command *Command) request(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, _ middlewares.NextFunc) {
	dimension, server, err := command.getOptions(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapPortalPositionRequest(dimension, server, lg)
	err = command.requestManager.Request(s, i, portalRequestRoutingKey, msg, command.respond)
	if err != nil {
		panic(err)
	}
}

func (command *Command) getOptions(ctx context.Context) (
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

func (command *Command) respond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	embeds := make([]*discordgo.MessageEmbed, 0)
	for _, position := range message.GetPortalPositionAnswer().GetPositions() {
		embeds = append(embeds, mappers.MapPortalToEmbed(position, command.portalService,
			command.serverService, message.Language))
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds})
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func isAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_PORTAL_POSITION_ANSWER &&
		message.PortalPositionAnswer != nil
}
