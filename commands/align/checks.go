package align

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

//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
func (command *Command) checkMandatoryCity(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] city
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.AlignCityOptionName {
				cities := command.bookService.FindCities(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetCitiesLabels(cities, i.Locale)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.city", option.StringValue(), labels, i.Locale)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyCity, cities[0]))
				} else {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("City check response ignored")
					}
				}

				return
			}
		}
	}
}

//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
func (command *Command) checkOptionalCity(ctx context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] city
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.AlignCityOptionName {
				cities := command.bookService.FindCities(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetCitiesLabels(cities, i.Locale)
				_, checkSuccess := validators.ExpectOnlyOneElement("checks.city", option.StringValue(), labels, i.Locale)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyCity, cities[0]))
				} else {
					next(ctx)
				}

				return
			}
		}
	}

	next(ctx)
}

//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
func (command *Command) checkMandatoryOrder(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] order
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.AlignOrderOptionName {
				orders := command.bookService.FindOrders(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetOrdersLabels(orders, i.Locale)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.order", option.StringValue(), labels, i.Locale)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyOrder, orders[0]))
				} else {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("Order check response ignored")
					}
				}

				return
			}
		}
	}
}

//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
func (command *Command) checkOptionalOrder(ctx context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] order
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.AlignOrderOptionName {
				orders := command.bookService.FindOrders(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetOrdersLabels(orders, i.Locale)
				_, checkSuccess := validators.ExpectOnlyOneElement("checks.order", option.StringValue(), labels, i.Locale)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyOrder, orders[0]))
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
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.AlignLevelOptionName {
				level := option.IntValue()

				if level >= constants.AlignmentMinLevel && level <= constants.AlignmentMaxLevel {
					next(context.WithValue(ctx, constants.ContextKeyLevel, level))
				} else {
					content := i18n.Get(i.Locale, "checks.level.constraints",
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
