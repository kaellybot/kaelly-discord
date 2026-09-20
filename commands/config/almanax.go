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

func (command *Command) almanaxRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	channelID, enabled, err := getWebhookAlmanaxOptions(ctx)
	if err != nil {
		return err
	}

	almanaxNews := command.almanaxService.GetAlmanaxNews(i.Locale)
	if almanaxNews == nil {
		// Discord locales the bot does not translate have no almanax channel of
		// their own: the default language is followed instead, as help promises.
		almanaxNews = command.almanaxService.GetAlmanaxNews(i18n.DefaultLocale)
	}

	if almanaxNews == nil {
		return fmt.Errorf("cannot find almanax news for locale %v", i.Locale)
	}

	var webhookID string
	if enabled {
		var created bool
		webhookID, created, err = command.followAnnouncement(s, i, almanaxNews.NewsChannelID, channelID)
		if err != nil {
			return err
		}

		if !created {
			return nil
		}
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapConfigurationNotificationRequest(i.GuildID, channelID, webhookID,
		authorID, "", amqp.NotificationType_ALMANAX, enabled, i.Locale)
	return command.requestManager.Request(s, i, constants.ConfigurationRequestRoutingKey,
		msg, command.setNotificationRespond)
}
