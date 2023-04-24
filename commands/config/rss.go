package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/mappers"
)

func (command *ConfigCommand) rssRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	channelId, feedId, enabled, locale, err := command.getWebhookRssOptions(ctx)
	if err != nil {
		panic(err)
	}

	webhook, err := command.createWebhook(s, channelId)
	if err != nil {
		panic(err)
		// TODO
	}

	msg := mappers.MapConfigurationWebhookRssRequest(webhook, feedId, enabled, locale, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.rssRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) rssRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	// TODO respond
}
