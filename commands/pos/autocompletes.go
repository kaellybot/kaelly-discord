package pos

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *PosCommand) autocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	data := i.ApplicationCommandData()
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	for _, option := range data.Options {
		if option.Focused {
			switch option.Name {
			case dimensionOptionName:
				dimensions := command.dimensionService.FindDimensions(option.StringValue(), lg)

				for _, dimension := range dimensions {
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:              dimension.Id,
						NameLocalizations: *i18n.GetLocalizations(dimension.Id),
						Value:             dimension.Id,
					})
				}
			case serverOptionName:
				servers := command.serverService.FindServers(option.StringValue(), lg)

				for _, server := range servers {
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:              server.Id,
						NameLocalizations: *i18n.GetLocalizations(server.Id),
						Value:             server.Id,
					})
				}
			default:
				log.Error().Str(constants.LogCommandOption, option.Name).Msgf("Option name not handled, ignoring it")
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
