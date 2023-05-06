package pos

import (
	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
)

func (command *Command) autocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	data := i.ApplicationCommandData()
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	for _, option := range data.Options {
		if option.Focused {
			switch option.Name {
			case contract.PosDimensionOptionName:
				dimensions := command.portalService.FindDimensions(option.StringValue(), lg)

				for _, dimension := range dimensions {
					label := translators.GetEntityLabel(dimension, lg)
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:  label,
						Value: label,
					})
				}
			case contract.PosServerOptionName:
				servers := command.serverService.FindServers(option.StringValue(), lg)

				for _, server := range servers {
					label := translators.GetEntityLabel(server, lg)
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:  label,
						Value: label,
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
