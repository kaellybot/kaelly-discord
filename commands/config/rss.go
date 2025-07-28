package config

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
)

func (command *Command) rssRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	channelID, feed, enabled, err := getWebhookRssOptions(ctx)
	if err != nil {
		panic(err)
	}

	var newsChannelID string
	for _, source := range feed.Sources {
		if source.Locale == i18n.MapDiscordLocale(i.Locale) {
			newsChannelID = source.NewsChannelID
			break
		}
	}

	if newsChannelID == "" {
		panic(fmt.Errorf("cannot find feed source for '%v' in %v", feed.ID, i.Locale))
	}

	var webhookID string
	if enabled {
		var created bool
		webhookID, created = command.followAnnouncement(s, i, newsChannelID, channelID)
		if !created {
			return
		}
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapConfigurationNotificationRequest(i.GuildID, channelID, webhookID,
		authorID, feed.ID, amqp.NotificationType_RSS, enabled, i.Locale)
	err = command.requestManager.Request(s, i, constants.ConfigurationRequestRoutingKey,
		msg, command.setNotificationRespond)
	if err != nil {
		panic(err)
	}
}
