package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *ConfigCommand) rssRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	channelId, feed, enabled, locale, err := command.getWebhookRssOptions(ctx)
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

	webhook, err := command.createWebhook(s, channelId)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapConfigurationWebhookRssRequest(webhook, feed, enabled, locale, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.rssRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) rssRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	// TODO respond
}
