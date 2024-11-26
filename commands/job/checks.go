package job

import (
	"context"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkJob(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] job
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.JobJobOptionName {
				jobs := command.bookService.FindJobs(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetJobsLabels(jobs, i.Locale)
				response, checkSuccess := validators.
					ExpectOnlyOneElement("checks.job", option.StringValue(), labels, i.Locale)
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
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.JobLevelOptionName {
				level := option.IntValue()

				if level >= constants.JobMinLevel && level <= constants.JobMaxLevel {
					next(context.WithValue(ctx, constants.ContextKeyLevel, level))
				} else {
					content := i18n.Get(i.Locale, "checks.level.constraints",
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
