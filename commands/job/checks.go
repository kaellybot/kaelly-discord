package job

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkJob(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] job
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == jobOptionName {
				jobs := command.bookService.FindJobs(option.StringValue(), lg)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.job", option.StringValue(), jobs, lg)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyJob, jobs[0]))
				} else {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("Job check response ignored")
					}
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

					if level >= constants.JobMinLevel && level <= constants.JobMaxLevel {
						next(context.WithValue(ctx, constants.ContextKeyLevel, level))
					} else {
						content := i18n.Get(lg, "checks.level.constraints",
							i18n.Vars{"min": constants.JobMinLevel, "max": constants.JobMaxLevel})
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

	// Option not filled (refers to guild and/or channel)
	server, found, err := command.guildService.GetServer(i.GuildID, i.ChannelID)
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
