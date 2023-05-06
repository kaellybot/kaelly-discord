package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkServer(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		if subCommand.Name == serverSubCommandName {
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
	}

	next(ctx)
}

func (command *Command) checkFeedType(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		if subCommand.Name == rssSubCommandName {
			for _, option := range subCommand.Options {
				if option.Name == feedTypeOptionName {
					feedTypes := command.feedService.FindFeedTypes(option.StringValue(), lg)
					response, checkSuccess := validators.ExpectOnlyOneElement("checks.feed", option.StringValue(), feedTypes, lg)
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
	}

	next(ctx)
}

func (command *Command) checkLanguage(ctx context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, _ discordgo.Locale, next middlewares.NextFunc) {
	locale := amqp.Language_ANY
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == languageOptionName {
				locale = amqp.Language(option.IntValue())
				break
			}
		}
	}

	next(context.WithValue(ctx, constants.ContextKeyLanguage, locale))
}

func (command *Command) checkChannelID(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == channelOptionName {
				next(context.WithValue(ctx, constants.ContextKeyChannel, option.ChannelValue(s).ID))
				return
			}
		}

		// If option not found, guess we're using the current channel for webhook queries
		if subCommand.Name != serverSubCommandName {
			next(context.WithValue(ctx, constants.ContextKeyChannel, i.ChannelID))
			return
		}
	}

	next(ctx)
}

func (command *Command) checkEnabled(ctx context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, _ discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == enabledOptionName {
				next(context.WithValue(ctx, constants.ContextKeyEnabled, option.BoolValue()))
				return
			}
		}
	}

	next(ctx)
}
