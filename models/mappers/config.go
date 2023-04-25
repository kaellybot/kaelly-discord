package mappers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/feeds"
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
	Language string
	Provider i18nProvider
}

type i18nProvider struct {
	Name  string
	Emoji string
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
		Type:     amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationSetServerRequest: &amqp.ConfigurationSetServerRequest{
			GuildId:   guildId,
			ChannelId: channelId,
			ServerId:  serverId,
		},
	}
}

func MapConfigurationWebhookAlmanaxRequest(webhook *discordgo.Webhook, guildId, channelId string,
	enabled bool, locale amqp.Language, lg discordgo.Locale) *amqp.RabbitMQMessage {

	if locale == amqp.Language_ANY {
		locale = constants.MapDiscordLocale(lg)
	}

	var webhookId, webhookToken string
	if webhook != nil {
		webhookId = webhook.ID
		webhookToken = webhook.Token
	}

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_SET_ALMANAX_WEBHOOK_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationSetAlmanaxWebhookRequest: &amqp.ConfigurationSetAlmanaxWebhookRequest{
			GuildId:      guildId,
			ChannelId:    channelId,
			WebhookId:    webhookId,
			WebhookToken: webhookToken,
			Enabled:      enabled,
			Language:     locale,
		},
	}
}

func MapConfigurationWebhookRssRequest(webhook *discordgo.Webhook, guildId, channelId string,
	feed entities.FeedType, enabled bool, locale amqp.Language, lg discordgo.Locale) *amqp.RabbitMQMessage {

	if locale == amqp.Language_ANY {
		locale = constants.MapDiscordLocale(lg)
	}

	var webhookId, webhookToken string
	if webhook != nil {
		webhookId = webhook.ID
		webhookToken = webhook.Token
	}

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_SET_RSS_WEBHOOK_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationSetRssWebhookRequest: &amqp.ConfigurationSetRssWebhookRequest{
			GuildId:      guildId,
			ChannelId:    channelId,
			FeedId:       feed.Id,
			WebhookId:    webhookId,
			WebhookToken: webhookToken,
			Enabled:      enabled,
			Language:     locale,
		},
	}
}

func MapConfigurationWebhookTwitterRequest(webhook *discordgo.Webhook, guildId, channelId string,
	enabled bool, locale amqp.Language, lg discordgo.Locale) *amqp.RabbitMQMessage {

	if locale == amqp.Language_ANY {
		locale = constants.MapDiscordLocale(lg)
	}

	var webhookId, webhookToken string
	if webhook != nil {
		webhookId = webhook.ID
		webhookToken = webhook.Token
	}

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_SET_TWITTER_WEBHOOK_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		ConfigurationSetTwitterWebhookRequest: &amqp.ConfigurationSetTwitterWebhookRequest{
			GuildId:      guildId,
			ChannelId:    channelId,
			WebhookId:    webhookId,
			WebhookToken: webhookToken,
			Enabled:      enabled,
			Language:     locale,
		},
	}
}

func MapConfigToEmbed(guild constants.GuildConfig, serverService servers.ServerService,
	feedService feeds.FeedService, locale amqp.Language) *discordgo.MessageEmbed {

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
	channelWebhooks = append(channelWebhooks, mapAlmanaxWebhooksToI18n(guild.AlmanaxWebhooks, lg)...)
	channelWebhooks = append(channelWebhooks, mapRssWebhooksToI18n(guild.RssWebhooks, feedService, lg)...)
	channelWebhooks = append(channelWebhooks, mapTwitterWebhooksToI18n(guild.TwitterWebhooks, lg)...)

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

func mapAlmanaxWebhooksToI18n(webhooks []constants.AlmanaxWebhook, lg discordgo.Locale) []i18nChannelWebhook {
	i18nWebhooks := make([]i18nChannelWebhook, 0)
	for _, webhook := range webhooks {
		i18nWebhooks = append(i18nWebhooks, i18nChannelWebhook{
			Channel: webhook.Channel.Mention(),
			Provider: i18nProvider{
				Name:  i18n.Get(lg, "webhooks.ALMANAX.name"),
				Emoji: i18n.Get(lg, "webhooks.ALMANAX.emoji"),
			},
			Language: i18n.Get(lg, fmt.Sprintf("locales.%s.emoji", webhook.Locale)),
		})
	}
	return i18nWebhooks
}

func mapRssWebhooksToI18n(webhooks []constants.RssWebhook, feedService feeds.FeedService, lg discordgo.Locale) []i18nChannelWebhook {
	i18nWebhooks := make([]i18nChannelWebhook, 0)
	for _, webhook := range webhooks {
		var providerName string
		feeds := feedService.FindFeedTypes(webhook.FeedId, lg)
		if len(feeds) == 1 {
			providerName = translators.GetEntityLabel(feeds[0], lg)
		} else {
			log.Warn().Str(constants.LogEntity, webhook.FeedId).
				Msgf("Cannot find feed type based on ID sent internally, continuing with default feed label")
			providerName = i18n.Get(lg, "webhooks.RSS.name")
		}

		i18nWebhooks = append(i18nWebhooks, i18nChannelWebhook{
			Channel: webhook.Channel.Mention(),
			Provider: i18nProvider{
				Name:  providerName,
				Emoji: i18n.Get(lg, "webhooks.RSS.emoji"),
			},
			Language: i18n.Get(lg, fmt.Sprintf("locales.%s.emoji", webhook.Locale)),
		})
	}
	return i18nWebhooks
}

func mapTwitterWebhooksToI18n(webhooks []constants.TwitterWebhook, lg discordgo.Locale) []i18nChannelWebhook {
	i18nWebhooks := make([]i18nChannelWebhook, 0)
	for _, webhook := range webhooks {
		i18nWebhooks = append(i18nWebhooks, i18nChannelWebhook{
			Channel: webhook.Channel.Mention(),
			Provider: i18nProvider{
				Name:  webhook.TwitterName,
				Emoji: i18n.Get(lg, "webhooks.TWITTER.emoji"),
			},
			Language: i18n.Get(lg, fmt.Sprintf("locales.%s.emoji", webhook.Locale)),
		})
	}
	return i18nWebhooks
}
