package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
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

func MapConfigToEmbed(guild constants.GuildConfig, serverService servers.ServerService,
	locale amqp.Language) *discordgo.MessageEmbed {

	lg := constants.MapAmqpLocale(locale)

	result := discordgo.MessageEmbed{
		Title: guild.Name,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: guild.Icon,
		},
		Color: constants.Color,
	}

	for _, channelServer := range guild.ChannelServers {
		server, found := serverService.GetServer(channelServer.ServerId)
		if !found {
			log.Warn().Str(constants.LogEntity, channelServer.ServerId).
				Msgf("Cannot find server based on ID sent internally, continuing with empty server")
			server = entities.Server{Id: channelServer.ServerId}
		}

		result.Fields = append(result.Fields, &discordgo.MessageEmbedField{
			Name:  channelServer.ChannelName,
			Value: server.Emoji + " " + translators.GetEntityLabel(server, lg),
		})
	}

	// TODO webhook

	return &result
}
