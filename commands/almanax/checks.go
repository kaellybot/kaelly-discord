package almanax

import (
	"context"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/parsing"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkDate(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	date := time.Now()
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.AlmanaxDateOptionName && len(strings.TrimSpace(option.StringValue())) > 0 {
				parsedDate, err := parsing.ParseDate(option.StringValue())
				if err != nil {
					content := i18n.Get(lg, "checks.date.constraints")
					_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
						Content: &content,
					})
					if err != nil {
						log.Error().Err(err).Msg("Date check response ignored")
					}
					return
				}
				date = *parsedDate
				break
			}
		}
	}

	if date.Before(constants.GetAlmanaxFirstDate()) || date.After(constants.GetAlmanaxLastDate()) {
		content := i18n.Get(lg, "checks.date.outOfBounds")
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Date check response ignored")
		}
		return
	}

	next(context.WithValue(ctx, constants.ContextKeyDate, &date))
}

func (command *Command) checkDuration(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.AlmanaxDurationOptionName {
				duration := int32(option.IntValue())

				if duration >= contract.AlmanaxDurationMinimumValue && duration <= contract.AlmanaxDurationMaximumValue {
					next(context.WithValue(ctx, constants.ContextKeyDuration, duration))
				} else {
					content := i18n.Get(lg, "checks.duration.constraints",
						i18n.Vars{"min": contract.AlmanaxDurationMinimumValue, "max": contract.AlmanaxDurationMaximumValue})
					_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
						Content: &content,
					})
					if err != nil {
						log.Error().Err(err).Msg("Duration check response ignored")
					}
				}

				return
			}
		}
	}

	next(context.WithValue(ctx, constants.ContextKeyDuration, int32(contract.AlmanaxDurationDefaultValue)))
}

func (command *Command) checkQuery(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.AlmanaxEffectOptionName && len(strings.TrimSpace(option.StringValue())) > 0 {
				next(context.WithValue(ctx, constants.ContextKeyQuery, option.StringValue()))
				return
			}
		}
	}

	content := i18n.Get(lg, "checks.query.constraints")
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		log.Error().Err(err).Msg("Query check response ignored")
	}
}
