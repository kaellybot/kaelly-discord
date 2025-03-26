package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	"github.com/rs/zerolog/log"
)

//nolint:dupl // OK for DRY concept but refactor at any cost is not relevant here.
func (command *Command) checkFeedType(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigFeedTypeOptionName {
				feedTypes := command.feedService.FindFeedTypes(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetFeedTypesLabels(feedTypes, i.Locale)
				response, checkSuccess := validators.
					ExpectOnlyOneElement("checks.feed", option.StringValue(), labels, i.Locale)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyFeed, feedTypes[0]))
				} else {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("Feed check response ignored")
					}
				}

				return
			}
		}
	}

	next(ctx)
}

//nolint:dupl // OK for DRY concept but refactor at any cost is not relevant here.
func (command *Command) checkTwitterAccount(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigTwitterAccountOptionName {
				twitterAccounts := command.twitterService.FindTwitterAccounts(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetTwittersLabels(twitterAccounts, i.Locale)
				response, checkSuccess := validators.
					ExpectOnlyOneElement("checks.twitterAccount", option.StringValue(), labels, i.Locale)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyTwitter, twitterAccounts[0]))
				} else {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("Twitter check response ignored")
					}
				}

				return
			}
		}
	}

	next(ctx)
}

func (command *Command) checkChannelID(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigChannelOptionName {
				next(context.WithValue(ctx, constants.ContextKeyChannel, option.ChannelValue(s).ID))
				return
			}
		}

		// If option not found, guess we're using the current channel for webhook queries
		if subCommand.Name != contract.ConfigServerSubCommandName {
			next(context.WithValue(ctx, constants.ContextKeyChannel, i.ChannelID))
			return
		}
	}

	next(ctx)
}

func (command *Command) checkEnabled(ctx context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigEnabledOptionName {
				next(context.WithValue(ctx, constants.ContextKeyEnabled, option.BoolValue()))
				return
			}
		}
	}

	next(ctx)
}
