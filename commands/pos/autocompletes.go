package pos

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

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
