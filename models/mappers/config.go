package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

func MapConfigurationDisplayRequest(guildId string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_DISPLAY_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationDisplayRequest: &amqp.ConfigurationDisplayRequest{
			GuildId: guildId,
		},
	}
}

func MapConfigurationServerRequest(guildId, channelId, serverId string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_SERVER_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationServerRequest: &amqp.ConfigurationServerRequest{
			GuildId:   guildId,
			ChannelId: channelId,
			ServerId:  serverId,
		},
	}
}

func MapConfigurationWebhookRequest(guildId, channelId string, enabled bool,
	provider amqp.ConfigurationWebhookRequest_Provider, lg discordgo.Locale) *amqp.RabbitMQMessage {

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_WEBHOOK_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationWebhookRequest: &amqp.ConfigurationWebhookRequest{
			GuildId:   guildId,
			ChannelId: channelId,
			Enabled:   enabled,
			Provider:  provider,
		},
	}
}
