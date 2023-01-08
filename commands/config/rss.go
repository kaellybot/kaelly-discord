package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/mappers"
)

func (command *ConfigCommand) rssRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	channelId, enabled, err := command.getWebhookOptions(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapConfigurationWebhookRequest(i.Interaction.GuildID, channelId,
		enabled, amqp.ConfigurationRequest_WebhookField_RSS, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.rssRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) rssRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage) {

	// TODO respond
}
