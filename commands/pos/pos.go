package pos

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	"github.com/kaellybot/kaelly-discord/services/dimensions"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func New(guildService guilds.GuildService, dimensionService dimensions.DimensionService, serverService servers.ServerService) *PosCommand {
	return &PosCommand{
		guildService:     guildService,
		dimensionService: dimensionService,
		serverService:    serverService,
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

func (command *PosCommand) respond(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle pos defer reponse, trying to continue")
	}

	// TODO send models to rabbitmq, retrieve response when possible

	data := make([]*discordgo.MessageEmbed, 0)
	if len(i.ApplicationCommandData().Options) > 0 {
		data = append(data, &discordgo.MessageEmbed{Title: i.ApplicationCommandData().Options[0].StringValue()})
	} else {
		data = append(data,
			&discordgo.MessageEmbed{Title: "ça"},
			&discordgo.MessageEmbed{Title: "fonctionne"},
			&discordgo.MessageEmbed{Title: "plutôt"},
			&discordgo.MessageEmbed{Title: "bien"},
		)
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &data,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle pos reponse")
	}
}
