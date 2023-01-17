package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

type i18nChannelServer struct {
	Channel string
	Server  i18nServer
}

type i18nServer struct {
	Name  string
	Emoji string
}

type i18nChannelWebhook struct {
	Channel  string
	Provider string
	Language string
	// TODO
}

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

	var guildServer *i18nServer
	if len(guild.ServerId) > 0 {
		server, found := serverService.GetServer(guild.ServerId)
		if !found {
			log.Warn().Str(constants.LogEntity, guild.ServerId).
				Msgf("Cannot find server based on ID sent internally, continuing with empty server")
			server = entities.Server{Id: guild.ServerId}
		}

		guildServer = &i18nServer{
			Name:  translators.GetEntityLabel(server, lg),
			Emoji: server.Emoji,
		}
	}

	channelServers := make([]i18nChannelServer, 0)
	for _, channelServer := range guild.ChannelServers {
		server, found := serverService.GetServer(channelServer.ServerId)
		if !found {
			log.Warn().Str(constants.LogEntity, channelServer.ServerId).
				Msgf("Cannot find server based on ID sent internally, continuing with empty server")
			server = entities.Server{Id: channelServer.ServerId}
		}

		channelServers = append(channelServers, i18nChannelServer{
			Channel: channelServer.Channel.Mention(),
			Server: i18nServer{
				Name:  translators.GetEntityLabel(server, lg),
				Emoji: server.Emoji,
			},
		})
	}

	channelWebhooks := make([]i18nChannelWebhook, 0)
	for _, channelWebhook := range guild.ChannelWebhooks {
		channelWebhooks = append(channelWebhooks, i18nChannelWebhook{
			Channel: channelWebhook.Channel.Mention(),
			// TODO webhook
		})
	}

	return &discordgo.MessageEmbed{
		Title:       guild.Name,
		Description: i18n.Get(lg, "config.embed.description", i18n.Vars{"server": guildServer, "game": constants.Game}),
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: guild.Icon},
		Color:       constants.Color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   i18n.Get(lg, "config.embed.server.name", i18n.Vars{"game": constants.Game}),
				Value:  i18n.Get(lg, "config.embed.server.value", i18n.Vars{"channels": channelServers}),
				Inline: true,
			},
			{
				Name:   i18n.Get(lg, "config.embed.webhook.name"),
				Value:  i18n.Get(lg, "config.embed.webhook.value", i18n.Vars{"channels": channelWebhooks}),
				Inline: true,
			},
		},
	}
}
