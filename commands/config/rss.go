//nolint:dupl // the code is duplicate but quite difficult to refactor: the needs behind are not the same
package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) rssRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	channelID, feed, enabled, err := getWebhookRssOptions(ctx)
	if err != nil {
		panic(err)
	}

	if !validators.HasWebhookPermission(s, channelID) {
		content := i18n.Get(i.Locale, "checks.permission.webhook")
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Permission check response ignored")
		}
		return
	}

	// TODO FOLLOW
	var webhookID string
	if enabled {
		var created bool
		webhookID, created = command.followAnnouncement(s, i, "1351966859779641406", channelID)
		if !created {
			return
		}
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapConfigurationRssRequest(i.GuildID, channelID, webhookID,
		authorID, feed, enabled, i.Locale)
	err = command.requestManager.Request(s, i, constants.ConfigurationRequestRoutingKey,
		msg, command.setRespond)
	if err != nil {
		panic(err)
	}
}
