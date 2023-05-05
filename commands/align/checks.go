package align

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkCity(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] city
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == cityOptionName {
				cities := command.bookService.FindCities(option.StringValue(), lg)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.city", option.StringValue(), cities, lg)
				if checkSuccess {
					next(context.WithValue(ctx, cityOptionName, cities[0]))
				} else if subCommand.Name == setSubCommandName {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("City check response ignored")
					}
				} else {
					next(ctx)
				}

				return
			}
		}
	}

	next(ctx)
}

func (command *Command) checkOrder(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] order
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == orderOptionName {
				orders := command.bookService.FindOrders(option.StringValue(), lg)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.order", option.StringValue(), orders, lg)
				if checkSuccess {
					next(context.WithValue(ctx, orderOptionName, orders[0]))
				} else if subCommand.Name == setSubCommandName {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("Order check response ignored")
					}
				} else {
					next(ctx)
				}

				return
			}
		}
	}

	next(ctx)
}

func (command *Command) checkLevel(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	for _, subCommand := range data.Options {
		if subCommand.Name == setSubCommandName {
			for _, option := range subCommand.Options {
				if option.Name == levelOptionName {
					level := option.IntValue()

					if level >= constants.AlignmentMinLevel && level <= constants.AlignmentMaxLevel {
						next(context.WithValue(ctx, levelOptionName, level))
					} else {
						content := i18n.Get(lg, "checks.level.constraints",
							i18n.Vars{"min": constants.AlignmentMinLevel, "max": constants.AlignmentMaxLevel})
						_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
							Content: &content,
						})
						if err != nil {
							log.Error().Err(err).Msg("Level check response ignored")
						}
					}

					return
				}
			}
		}
	}

	next(ctx)
}

func (command *Command) checkServer(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] server
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == serverOptionName {
				servers := command.serverService.FindServers(option.StringValue(), lg)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.server", option.StringValue(), servers, lg)
				if checkSuccess {
					next(context.WithValue(ctx, serverOptionName, servers[0]))
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

	// Option not filled (refers to guild and/or channel)
	server, err := command.guildService.GetServer(i.GuildID, i.ChannelID)
	if err != nil {
		panic(err)
	}

	if server == nil {
		content := i18n.Get(lg, "checks.server.required", i18n.Vars{"game": constants.Game})
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Server check response ignored")
		}
	} else {
		next(context.WithValue(ctx, serverOptionName, *server))
	}
}
