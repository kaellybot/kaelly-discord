package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkFeedType(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigFeedTypeOptionName {
				feedTypes := command.feedService.FindFeedTypes(option.StringValue(), i.Locale)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.feed", option.StringValue(), feedTypes, i.Locale)
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

func (command *Command) checkVideast(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigVideastOptionName {
				videasts := command.videastService.FindVideasts(option.StringValue(), i.Locale)
				response, checkSuccess := validators.ExpectOnlyOneElement("checks.videast", option.StringValue(), videasts, i.Locale)
				if checkSuccess {
					next(context.WithValue(ctx, constants.ContextKeyVideast, videasts[0]))
				} else {
					_, err := s.InteractionResponseEdit(i.Interaction, &response)
					if err != nil {
						log.Error().Err(err).Msg("Videast check response ignored")
					}
				}

				return
			}
		}
	}

	next(ctx)
}

func (command *Command) checkLanguage(ctx context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	locale := amqp.Language_ANY
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigLanguageOptionName {
				locale = amqp.Language(option.IntValue())
				break
			}
		}
	}

	next(context.WithValue(ctx, constants.ContextKeyLanguage, locale))
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
