package align

import (
	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
)

func (command *Command) autocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Focused {
				switch option.Name {
				case contract.AlignCityOptionName:
					choices = command.findCities(option.StringValue(), lg)
				case contract.AlignOrderOptionName:
					choices = command.findOrders(option.StringValue(), lg)
				case contract.AlignServerOptionName:
					choices = command.findServers(option.StringValue(), lg)
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

func (command *Command) findCities(cityName string, lg discordgo.Locale) []*discordgo.
	ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	cities := command.bookService.FindCities(cityName, lg)

	for _, city := range cities {
		label := translators.GetEntityLabel(city, lg)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  label,
			Value: label,
		})
	}

	return choices
}

func (command *Command) findOrders(orderName string, lg discordgo.Locale) []*discordgo.
	ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	orders := command.bookService.FindOrders(orderName, lg)

	for _, order := range orders {
		label := translators.GetEntityLabel(order, lg)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  label,
			Value: label,
		})
	}

	return choices
}
