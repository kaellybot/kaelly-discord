package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *ConfigCommand) checkServer(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		if subCommand.Name == serverSubCommandName {
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
	}

	next(ctx)
}

func (command *ConfigCommand) checkFeedType(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		if subCommand.Name == rssSubCommandName {
			for _, option := range subCommand.Options {
				if option.Name == feedTypeOptionName {
					feedTypes := command.feedService.FindFeedTypes(option.StringValue(), lg)
					response, checkSuccess := validators.ExpectOnlyOneElement("checks.feed", option.StringValue(), feedTypes, lg)
					if checkSuccess {
						next(context.WithValue(ctx, feedTypeOptionName, feedTypes[0]))
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

func (command *ConfigCommand) checkLanguage(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

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

	next(context.WithValue(ctx, languageOptionName, locale))
}

func (command *ConfigCommand) checkChannelId(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == channelOptionName {
				next(context.WithValue(ctx, channelOptionName, option.ChannelValue(s).ID))
				return
			}
		}

		// If option not found, guess we're using the current channel for webhook queries
		if subCommand.Name != serverSubCommandName {
			next(context.WithValue(ctx, channelOptionName, i.ChannelID))
			return
		}
	}

	next(ctx)
}

func (command *ConfigCommand) checkEnabled(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == enabledOptionName {
				next(context.WithValue(ctx, enabledOptionName, option.BoolValue()))
				return
			}
		}
	}

	next(ctx)
}

func (command *ConfigCommand) checkWebhookPermission(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	// TODO l'id du channel n'est pas forcément le bon ;)

	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		if subCommand.Name == almanaxSubCommandName ||
			subCommand.Name == rssSubCommandName ||
			subCommand.Name == twitterSubCommandName {

			permissions, err := s.State.UserChannelPermissions(s.State.User.ID, i.ChannelID)
			if err != nil {
				log.Error().Err(err).Msg("Cannot retrieve channel permission, check failed")
				content := i18n.Get(lg, "checks.permissions.webhook")
				_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
				if err != nil {
					log.Error().Err(err).Msg("Permission check response ignored")
				}
				return
			}
			if permissions&discordgo.PermissionManageWebhooks == 0 {
				content := i18n.Get(lg, "checks.permissions.webhook")
				_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
				if err != nil {
					log.Error().Err(err).Msg("Permission check response ignored")
				}
				return
			}
		}
	}

	next(ctx)
}
