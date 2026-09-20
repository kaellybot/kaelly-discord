package config

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

//nolint:dupl // OK for DRY concept but refactor at any cost is not relevant here.
func (command *Command) checkFeedType(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) error {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigFeedTypeOptionName {
				feedTypes := command.feedService.FindFeedTypes(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetFeedTypesLabels(feedTypes, i.Locale)
				response, checkSuccess := validators.
					ExpectOnlyOneElement("checks.feed", option.StringValue(), labels, i.Locale)
				if checkSuccess {
					return next(context.WithValue(ctx, constants.ContextKeyFeed, feedTypes[0]))
				}

				_, err := s.InteractionResponseEdit(i.Interaction, &response)
				if err != nil {
					log.Error().Err(err).Msg("Feed check response ignored")
				}

				return nil
			}
		}
	}

	return next(ctx)
}

//nolint:dupl // OK for DRY concept but refactor at any cost is not relevant here.
func (command *Command) checkTwitterAccount(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) error {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigTwitterAccountOptionName {
				twitterAccounts := command.twitterService.FindTwitterAccounts(option.StringValue(), i.Locale, constants.MaxChoices)
				labels := translators.GetTwittersLabels(twitterAccounts, i.Locale)
				response, checkSuccess := validators.
					ExpectOnlyOneElement("checks.twitterAccount", option.StringValue(), labels, i.Locale)
				if checkSuccess {
					return next(context.WithValue(ctx, constants.ContextKeyTwitter, twitterAccounts[0]))
				}

				_, err := s.InteractionResponseEdit(i.Interaction, &response)
				if err != nil {
					log.Error().Err(err).Msg("Twitter check response ignored")
				}

				return nil
			}
		}
	}

	return next(ctx)
}

// checkServerChannelID binds the channel a thread belongs to: a thread has no
// configuration of its own, and that is how the binding is read back.
func (command *Command) checkServerChannelID(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) error {
	return command.checkChannelID(ctx, s, i, func(channelCtx context.Context) error {
		channelID, ok := channelCtx.Value(constants.ContextKeyChannel).(string)
		if !ok {
			return next(channelCtx)
		}

		return next(context.WithValue(channelCtx, constants.ContextKeyChannel,
			discord.ResolveParentChannelID(s, channelID)))
	})
}

func (command *Command) checkChannelID(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) error {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigChannelOptionName {
				return next(context.WithValue(ctx, constants.ContextKeyChannel, option.ChannelValue(s).ID))
			}
		}

		// If option not found, guess we're using the current channel for webhook queries
		if subCommand.Name != contract.ConfigServerSubCommandName {
			return next(context.WithValue(ctx, constants.ContextKeyChannel, i.ChannelID))
		}
	}

	return next(ctx)
}

func (command *Command) checkFollowConstraints(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) error {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		// Previous checks always fill it; fallback on the channel hosting the interaction.
		channelID = i.ChannelID
	}

	if channelID == "" {
		return fmt.Errorf("cannot retrieve channelID from ctx (%v)", constants.ContextKeyChannel)
	}

	// Only regular Text channels (GUILD_TEXT) could follow Announcement channels.
	channel, err := s.State.Channel(channelID)
	if err != nil {
		// Not cached yet, ask Discord about it.
		channel, err = s.Channel(channelID)
		if err != nil {
			return fmt.Errorf("%w: %w", commands.ErrUnavailable, err)
		}
	}

	if channel.Type != discordgo.ChannelTypeGuildText {
		content := i18n.Get(i.Locale, "checks.webhook.targetChannelType")
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Channel type check response ignored")
		}

		return nil
	}

	// But have we got the permission to follow?
	if !discord.HasPermissions(s, channelID, discordgo.PermissionManageWebhooks) {
		content := i18n.Get(i.Locale, "checks.webhook.permission")
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Permission check response ignored")
		}

		return nil
	}

	return next(ctx)
}

func (command *Command) checkEnabled(ctx context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) error {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.ConfigEnabledOptionName {
				return next(context.WithValue(ctx, constants.ContextKeyEnabled, option.BoolValue()))
			}
		}
	}

	return next(ctx)
}
