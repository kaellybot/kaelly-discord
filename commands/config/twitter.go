//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
)

func (command *Command) twitterRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	channelID, twitterAccount, enabled, err := getWebhookTwitterOptions(ctx)
	if err != nil {
		panic(err)
	}

	var webhookID string
	if enabled {
		var created bool
		webhookID, created = command.followAnnouncement(s, i, twitterAccount.NewsChannelID, channelID)
		if !created {
			return
		}
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapConfigurationNotificationRequest(i.GuildID, channelID, webhookID,
		authorID, twitterAccount.ID, amqp.NotificationType_TWITTER, enabled, i.Locale)
	err = command.requestManager.Request(s, i, constants.ConfigurationRequestRoutingKey,
		msg, command.setNotificationRespond)
	if err != nil {
		panic(err)
	}
}
