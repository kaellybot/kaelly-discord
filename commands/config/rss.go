package config

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
)

func (command *Command) rssRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	channelID, feed, enabled, err := getWebhookRssOptions(ctx)
	if err != nil {
		return err
	}

	// The default language is followed for Discord locales the bot does not
	// translate, which have no source of their own.
	newsChannelID := findNewsChannelID(feed, i.Locale)
	if newsChannelID == "" {
		newsChannelID = findNewsChannelID(feed, i18n.DefaultLocale)
	}

	if newsChannelID == "" {
		return fmt.Errorf("cannot find feed source for '%v' in %v", feed.ID, i.Locale)
	}

	var webhookID string
	if enabled {
		var created bool
		webhookID, created, err = command.followAnnouncement(s, i, newsChannelID, channelID)
		if err != nil {
			return err
		}

		if !created {
			return nil
		}
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapConfigurationNotificationRequest(i.GuildID, channelID, webhookID,
		authorID, feed.ID, amqp.NotificationType_RSS, enabled, i.Locale)
	return command.requestManager.Request(s, i, constants.ConfigurationRequestRoutingKey,
		msg, command.setNotificationRespond)
}

func findNewsChannelID(feed entities.FeedType, locale discordgo.Locale) string {
	for _, source := range feed.Sources {
		if source.Locale == i18n.MapDiscordLocale(locale) {
			return source.NewsChannelID
		}
	}

	return ""
}
