package pos

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models"
	"github.com/kaellybot/kaelly-discord/services/dimensions"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	i18n "github.com/kaysoro/discordgo-i18n"
)

const (
	portalRequestRoutingKey = "requests.portals"
)

func New(guildService guilds.GuildService, dimensionService dimensions.DimensionService,
	serverService servers.ServerService, broker amqp.MessageBrokerInterface) *PosCommand {

	return &PosCommand{
		guildService:     guildService,
		dimensionService: dimensionService,
		serverService:    serverService,
		broker:           broker,
	}
}

func (command *PosCommand) GetDiscordCommand() *models.DiscordCommand {
	return &models.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     commandName,
			Description:              i18n.Get(models.DefaultLocale, "pos.description"),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: &models.DefaultPermission,
			DMPermission:             &models.DMPermission,
			NameLocalizations:        i18n.GetLocalizations("pos.name"),
			DescriptionLocalizations: i18n.GetLocalizations("pos.description"),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:                     dimensionOptionName,
					Description:              i18n.Get(models.DefaultLocale, "pos.dimension.description"),
					NameLocalizations:        *i18n.GetLocalizations("pos.dimension.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("pos.dimension.description"),
					Type:                     discordgo.ApplicationCommandOptionString,
					Required:                 false,
					Autocomplete:             true,
				},
				{
					Name:                     serverOptionName,
					Description:              i18n.Get(models.DefaultLocale, "pos.server.description", i18n.Vars{"game": models.Game}),
					NameLocalizations:        *i18n.GetLocalizations("pos.server.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("pos.server.description", i18n.Vars{"game": models.Game}),
					Type:                     discordgo.ApplicationCommandOptionString,
					Required:                 false,
					Autocomplete:             true,
				},
			},
		},
		Handlers: models.DiscordHandlers{
			discordgo.InteractionApplicationCommand:             middlewares.Use(command.checkDimension, command.checkServer, command.respond),
			discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
		},
	}
}

func (command *PosCommand) respond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	dimension, server, err := command.getOptions(ctx)
	if err != nil {
		panic(err)
	}

	err = command.publishPortalPositionRequest(i.ID, dimension, server, lg)
	if err != nil {
		panic(err)
	}
	// TODO Add message in request manager
}

func (command *PosCommand) getOptions(ctx context.Context) (models.Dimension, models.Server, error) {
	server, ok := ctx.Value(serverOptionName).(models.Server)
	if !ok {
		return models.Dimension{}, models.Server{}, fmt.Errorf("Cannot cast %v as models.Server", ctx.Value(serverOptionName))
	}

	dimension := models.Dimension{}
	if ctx.Value(dimensionOptionName) != nil {
		dimension, ok = ctx.Value(dimensionOptionName).(models.Dimension)
		if !ok {
			return models.Dimension{}, models.Server{}, fmt.Errorf("Cannot cast %v as models.Dimension", ctx.Value(dimensionOptionName))
		}
	}

	return dimension, server, nil
}

func (command *PosCommand) publishPortalPositionRequest(id string, dimension models.Dimension,
	server models.Server, lg discordgo.Locale) error {
	msg := models.MapPortalPositionRequest(dimension, server, lg)
	return command.broker.Publish(msg, amqp.ExchangeRequest, portalRequestRoutingKey, id)
}
