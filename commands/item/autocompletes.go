package item

import (
	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/rs/zerolog/log"
)

func (command *Command) autocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Focused {
				switch option.Name {
				case contract.ItemQueryOptionName:
					choices = command.findItems(option.StringValue(), lg)
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

func (command *Command) findItems(serverName string, lg discordgo.Locale) []*discordgo.
	ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	// TODO

	return choices
}
