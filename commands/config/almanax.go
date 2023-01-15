package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/mappers"
)

func (command *ConfigCommand) almanaxRequest(ctx context.Context, s *discordgo.Session,
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
		enabled, amqp.ConfigurationSetRequest_WebhookField_ALMANAX, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.almanaxRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) almanaxRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	// TODO respond
}
