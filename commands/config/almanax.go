package config

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
)

func (command *Command) almanaxRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	channelID, enabled, err := getWebhookAlmanaxOptions(ctx)
	if err != nil {
		panic(err)
	}

	almanaxNews := command.almanaxService.GetAlmanaxNews(i.Locale)
	if almanaxNews == nil {
		panic(fmt.Errorf("cannot find almanax news for locale %v", i.Locale))
	}

	var webhookID string
	if enabled {
		var created bool
		webhookID, created = command.followAnnouncement(s, i, almanaxNews.NewsChannelID, channelID)
		if !created {
			return
		}
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapConfigurationNotificationRequest(i.GuildID, channelID, webhookID,
		authorID, "", amqp.NotificationType_ALMANAX, enabled, i.Locale)
	err = command.requestManager.Request(s, i, constants.ConfigurationRequestRoutingKey,
		msg, command.setNotificationRespond)
	if err != nil {
		panic(err)
	}
}
