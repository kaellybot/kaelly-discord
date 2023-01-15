package pos

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func New(guildService guilds.GuildService, portalService portals.PortalService,
	serverService servers.ServerService, requestManager requests.RequestManager) *PosCommand {

	return &PosCommand{
		guildService:   guildService,
		portalService:  portalService,
		serverService:  serverService,
		requestManager: requestManager,
	}
}

func (command *PosCommand) GetDiscordCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     commandName,
			Description:              i18n.Get(constants.DefaultLocale, "pos.description"),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: &constants.DefaultPermission,
			DMPermission:             &constants.DMPermission,
			NameLocalizations:        i18n.GetLocalizations("pos.name"),
			DescriptionLocalizations: i18n.GetLocalizations("pos.description"),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:                     dimensionOptionName,
					Description:              i18n.Get(constants.DefaultLocale, "pos.dimension.description"),
					NameLocalizations:        *i18n.GetLocalizations("pos.dimension.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("pos.dimension.description"),
					Type:                     discordgo.ApplicationCommandOptionString,
					Required:                 false,
					Autocomplete:             true,
				},
				{
					Name:                     serverOptionName,
					Description:              i18n.Get(constants.DefaultLocale, "pos.server.description", i18n.Vars{"game": constants.Game}),
					NameLocalizations:        *i18n.GetLocalizations("pos.server.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("pos.server.description", i18n.Vars{"game": constants.Game}),
					Type:                     discordgo.ApplicationCommandOptionString,
					Required:                 false,
					Autocomplete:             true,
				},
			},
		},
		Handlers: constants.DiscordHandlers{
			discordgo.InteractionApplicationCommand:             middlewares.Use(command.checkDimension, command.checkServer, command.request),
			discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
		},
	}
}

func (command *PosCommand) request(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

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

func (command *PosCommand) getOptions(ctx context.Context) (entities.Dimension, entities.Server, error) {
	server, ok := ctx.Value(serverOptionName).(entities.Server)
	if !ok {
		return entities.Dimension{}, entities.Server{}, fmt.Errorf("Cannot cast %v as entities.Server", ctx.Value(serverOptionName))
	}

	dimension := entities.Dimension{}
	if ctx.Value(dimensionOptionName) != nil {
		dimension, ok = ctx.Value(dimensionOptionName).(entities.Dimension)
		if !ok {
			return entities.Dimension{}, entities.Server{}, fmt.Errorf("Cannot cast %v as entities.Dimension", ctx.Value(dimensionOptionName))
		}
	}

	return dimension, server, nil
}

func (command *PosCommand) respond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	if !isAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	embeds := make([]*discordgo.MessageEmbed, 0)
	for _, position := range message.GetPortalPositionAnswer().GetPositions() {
		embeds = append(embeds, mappers.MapToEmbed(position, command.portalService, message.Language))
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds})
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func isAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_PORTAL_POSITION_ANSWER &&
		message.PortalPositionAnswer != nil
}
