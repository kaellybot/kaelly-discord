package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

func MapConfigurationGetRequest(guildId string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_GET_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationGetRequest: &amqp.ConfigurationGetRequest{
			GuildId: guildId,
		},
	}
}

func MapConfigurationServerRequest(guildId, channelId, serverId string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_SET_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationSetRequest: &amqp.ConfigurationSetRequest{
			GuildId:   guildId,
			ChannelId: channelId,
			Field:     amqp.ConfigurationSetRequest_SERVER,
			ServerField: &amqp.ConfigurationSetRequest_ServerField{
				ServerId: serverId,
			},
		},
	}
}

func MapConfigurationWebhookRequest(guildId, channelId string, enabled bool,
	provider amqp.ConfigurationSetRequest_WebhookField_Provider, lg discordgo.Locale) *amqp.RabbitMQMessage {

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_SET_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationSetRequest: &amqp.ConfigurationSetRequest{
			GuildId:   guildId,
			ChannelId: channelId,
			Field:     amqp.ConfigurationSetRequest_WEBHOOK,
			WebhookField: &amqp.ConfigurationSetRequest_WebhookField{
				// TODO Webhook id & token
				WebhookId:    "",
				WebhookToken: "",
				Enabled:      enabled,
				Provider:     provider,
			},
		},
	}
}
