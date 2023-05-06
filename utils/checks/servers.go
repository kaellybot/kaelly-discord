package checks

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func CheckServer(optionName string, serverService servers.Service) middlewares.MiddlewareCommand {
	return func(ctx context.Context, s *discordgo.Session,
		i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
		data := i.ApplicationCommandData()
		for _, subCommand := range data.Options {
			for _, option := range subCommand.Options {
				if option.Name == optionName {
					servers := serverService.FindServers(option.StringValue(), lg)
					response, checkSuccess := validators.ExpectOnlyOneElement("checks.server", option.StringValue(), servers, lg)
					if checkSuccess {
						next(context.WithValue(ctx, constants.ContextKeyServer, servers[0]))
					} else {
						_, err := s.InteractionResponseEdit(i.Interaction, &response)
						if err != nil {
							log.Error().Err(err).Msg("Server check response ignored")
						}
					}

					return
				}
			}
		}

		next(ctx)
	}
}

func CheckServerWithFallback(optionName string, serverService servers.Service,
	guildService guilds.Service) middlewares.MiddlewareCommand {
	return func(ctx context.Context, s *discordgo.Session,
		i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
		data := i.ApplicationCommandData()

		// Filled case, expecting [1, 1] server
		for _, option := range data.Options {
			serverValue, found := getServerValue(optionName, option)
			if found {
				servers := serverService.FindServers(serverValue, lg)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.server", serverValue, servers, lg)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyServer, servers[0]))
				} else {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("Server check response ignored")
					}
				}

				return
			}
		}

		// Option not filled (refers to guild and/or channel)
		tryFallback(ctx, s, i, lg, next, guildService)
	}
}

func getServerValue(optionName string, option *discordgo.ApplicationCommandInteractionDataOption) (string, bool) {
	if option.Type == discordgo.ApplicationCommandOptionString && option.Name == optionName {
		return option.StringValue(), true
	}

	for _, subOption := range option.Options {
		value, found := getServerValue(optionName, subOption)
		if found {
			return value, true
		}
	}

	return "", false
}

func tryFallback(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
	lg discordgo.Locale, next middlewares.NextFunc, guildService guilds.Service) {
	server, found, err := guildService.GetServer(i.GuildID, i.ChannelID)
	if err != nil {
		panic(err)
	}

	if !found {
		content := i18n.Get(lg, "checks.server.required", i18n.Vars{"game": constants.GetGame()})
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Server check response ignored")
		}
	} else {
		next(context.WithValue(ctx, constants.ContextKeyServer, server))
	}
}
