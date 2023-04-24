package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/mappers"
)

func (command *ConfigCommand) twitterRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	channelId, enabled, locale, err := command.getWebhookTwitterOptions(ctx)
	if err != nil {
		panic(err)
	}

	webhook, err := command.createWebhook(s, channelId)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapConfigurationWebhookTwitterRequest(webhook, enabled, locale, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.twitterRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) twitterRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	// TODO respond
}
