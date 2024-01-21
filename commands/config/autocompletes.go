package config

import (
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
)

func (command *Command) autocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Focused {
				switch option.Name {
				case contract.ConfigServerOptionName:
					choices = command.findServers(option.StringValue(), i.Locale)
				case contract.ConfigFeedTypeOptionName:
					choices = command.findFeedTypes(option.StringValue(), i.Locale)
				case contract.ConfigVideastOptionName:
					choices = command.findVideasts(option.StringValue(), i.Locale)
				case contract.ConfigStreamerOptionName:
					choices = command.findStreamers(option.StringValue(), i.Locale)
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

	sort.SliceStable(choices, func(i, j int) bool {
		return choices[i].Name < choices[j].Name
	})

	return choices
}

func (command *Command) findFeedTypes(feedTypeName string, lg discordgo.Locale) []*discordgo.
	ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	feedTypes := command.feedService.FindFeedTypes(feedTypeName, lg)

	for _, feedType := range feedTypes {
		label := translators.GetEntityLabel(feedType, lg)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  label,
			Value: label,
		})
	}

	sort.SliceStable(choices, func(i, j int) bool {
		return choices[i].Name < choices[j].Name
	})

	return choices
}

func (command *Command) findVideasts(videastName string, lg discordgo.Locale) []*discordgo.
	ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	if len(strings.TrimSpace(videastName)) == 0 {
		return choices
	}

	videasts := command.videastService.FindVideasts(videastName, lg)
	for _, videast := range videasts {
		label := translators.GetEntityLabel(videast, lg)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  label,
			Value: label,
		})
	}

	sort.SliceStable(choices, func(i, j int) bool {
		return choices[i].Name < choices[j].Name
	})

	return choices
}

func (command *Command) findStreamers(streamerName string, lg discordgo.Locale) []*discordgo.
	ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	if len(strings.TrimSpace(streamerName)) == 0 {
		return choices
	}

	streamers := command.streamerService.FindStreamers(streamerName, lg)
	for _, streamer := range streamers {
		label := translators.GetEntityLabel(streamer, lg)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  label,
			Value: label,
		})
	}

	sort.SliceStable(choices, func(i, j int) bool {
		return choices[i].Name < choices[j].Name
	})

	return choices
}
