package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *ConfigCommand) almanaxRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	channelId, enabled, locale, err := command.getWebhookAlmanaxOptions(ctx)
	if err != nil {
		panic(err)
	}

	if !validators.HasWebhookPermission(s, channelId) {
		content := i18n.Get(lg, "checks.permissions.webhook")
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Permission check response ignored")
		}
		return
	}

	var webhook *discordgo.Webhook = nil
	if enabled {
		webhook, err = command.createWebhook(s, channelId)
		if err != nil {
			panic(err)
		}
	}

	msg := mappers.MapConfigurationWebhookAlmanaxRequest(webhook, i.GuildID, channelId, enabled, locale, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.setRespond)
	if err != nil {
		panic(err)
	}
}
