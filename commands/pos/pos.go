package pos

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	"github.com/kaellybot/kaelly-discord/services/dimension"
	"github.com/kaellybot/kaelly-discord/services/server"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

const (
	commandName         = "pos"
	dimensionOptionName = "dimension"
	serverOptionName    = "server"
)

type PosCommand struct {
	dimensionService dimension.DimensionService
	serverService    server.ServerService
}

func New(dimensionService dimension.DimensionService, serverService server.ServerService) *PosCommand {
	return &PosCommand{
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
					Description:              i18n.Get(models.DefaultLocale, "pos.server.description"),
					NameLocalizations:        *i18n.GetLocalizations("pos.server.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("pos.server.description", i18n.Vars{"game": models.Game}),
					Type:                     discordgo.ApplicationCommandOptionString,
					Required:                 false,
					Autocomplete:             true,
				},
			},
		},
		Handler: command.handler,
	}
}

func (command *PosCommand) handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		command.respond(s, i)
	case discordgo.InteractionApplicationCommandAutocomplete:
		command.autocomplete(s, i)
	default:
		log.Error().Uint32(models.LogInteractionType, uint32(i.Type)).Msgf("Interaction not handled, ignoring it")
	}
}

func (command *PosCommand) respond(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

func (command *PosCommand) autocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	for _, option := range data.Options {
		if option.Focused {
			switch option.Name {
			case dimensionOptionName:
				dimensions := command.dimensionService.FindDimensions(option.StringValue(), i.Locale)

				for _, dimension := range dimensions {
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:              dimension.Name,
						NameLocalizations: *i18n.GetLocalizations(dimension.Name),
						Value:             dimension.Name,
					})
				}
			case serverOptionName:
				servers := command.serverService.FindServers(option.StringValue(), i.Locale)

				for _, server := range servers {
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:              server.Name,
						NameLocalizations: *i18n.GetLocalizations(server.Name),
						Value:             server.Name,
					})
				}
			default:
				log.Error().Str(models.LogCommandOption, option.Name).Msgf("Option name not handled, ignoring it")
			}
			break
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Autocomplete request ignored")
	}
}
