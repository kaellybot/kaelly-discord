package job

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
)

func (command *JobCommand) autocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	data := i.ApplicationCommandData()
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Focused {
				switch option.Name {
				case jobOptionName:
					jobs := command.bookService.FindJobs(option.StringValue(), lg)

					for _, job := range jobs {
						label := translators.GetEntityLabel(job, lg)
						choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
							Name:  label,
							Value: label,
						})
					}
				case serverOptionName:
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
