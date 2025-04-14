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
	"github.com/kaellybot/kaelly-discord/utils/validators"
	di18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) rssRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	channelID, feed, enabled, err := getWebhookRssOptions(ctx)
	if err != nil {
		panic(err)
	}

	if !validators.HasWebhookPermission(s, channelID) {
		content := di18n.Get(i.Locale, "checks.permission.webhook")
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Permission check response ignored")
		}
		return
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
