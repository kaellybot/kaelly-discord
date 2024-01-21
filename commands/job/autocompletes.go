package job

import (
	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
)

func (command *Command) autocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Focused {
				switch option.Name {
				case contract.JobJobOptionName:
					choices = command.findJobs(option.StringValue(), i.Locale)
				case contract.JobServerOptionName:
					choices = command.findServers(option.StringValue(), i.Locale)
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

func (command *Command) findJobs(jobName string, lg discordgo.Locale) []*discordgo.
	ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	jobs := command.bookService.FindJobs(jobName, lg)

	for _, job := range jobs {
		label := translators.GetEntityLabel(job, lg)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  label,
			Value: label,
		})
	}

	return choices
}

func (command *Command) findServers(serverName string, lg discordgo.Locale) []*discordgo.
	ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	servers := command.serverService.FindServers(serverName, lg)

	for _, server := range servers {
		label := translators.GetEntityLabel(server, lg)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  label,
			Value: label,
		})
	}

	return choices
}
