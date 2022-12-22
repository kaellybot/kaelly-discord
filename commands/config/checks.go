package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	"github.com/rs/zerolog/log"
)

func (command *ConfigCommand) checkServer(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] server
	for _, subCommand := range data.Options {
		if subCommand.Name == serverSubCommandName {
			for _, option := range subCommand.Options {
				if option.Name == serverOptionName {
					servers := command.serverService.FindServers(option.StringValue(), lg)
					response, checkSuccess := validators.ExpectOnlyOneElement("config.server.check", option.StringValue(), servers, lg)
					if checkSuccess {
						next(context.WithValue(ctx, serverOptionName, servers[0]))
					} else {
						err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &response,
						})
						if err != nil {
							log.Error().Err(err).Msg("Server check response ignored")
						}
					}

					return
				}
			}
		}
	}

	next(ctx)
}
