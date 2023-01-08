package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

func MapConfigurationServerRequest(guildId, channelId, serverId string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationRequest: &amqp.ConfigurationRequest{
			GuildId:   guildId,
			ChannelId: channelId,
			Field:     amqp.ConfigurationRequest_SERVER,
			ServerField: &amqp.ConfigurationRequest_ServerField{
				ServerId: serverId,
			},
		},
	}
}

func MapConfigurationWebhookRequest(guildId, channelId string, enabled bool,
	provider amqp.ConfigurationRequest_WebhookField_Provider, lg discordgo.Locale) *amqp.RabbitMQMessage {

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationRequest: &amqp.ConfigurationRequest{
			GuildId:   guildId,
			ChannelId: channelId,
			Field:     amqp.ConfigurationRequest_WEBHOOK,
			WebhookField: &amqp.ConfigurationRequest_WebhookField{
				WebhookId:    "TODO",
				WebhookToken: "TODO",
				Enabled:      enabled,
				Provider:     provider,
			},
		},
	}
}
