package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/twitters"
	"github.com/kaellybot/kaelly-discord/utils/discord"
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
	Provider i18nProvider
}

type i18nProvider struct {
	Name  string
	Emoji string
}

func MapConfigurationGetRequest(guildID, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_CONFIGURATION_GET_REQUEST, lg)
	request.ConfigurationGetRequest = &amqp.ConfigurationGetRequest{
		GuildId: guildID,
	}
	return request
}

func MapConfigurationServerRequest(guildID, channelID, serverID, authorID string,
	lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_REQUEST, lg)
	request.ConfigurationSetServerRequest = &amqp.ConfigurationSetServerRequest{
		GuildId:   guildID,
		ChannelId: channelID,
		ServerId:  serverID,
	}
	return request
}

func MapConfigurationAlmanaxRequest(guildID, channelID, webhookID, authorID string,
	enabled bool, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_CONFIGURATION_SET_ALMANAX_REQUEST, lg)

	request.ConfigurationSetAlmanaxRequest = &amqp.ConfigurationSetAlmanaxRequest{
		GuildId:   guildID,
		ChannelId: channelID,
		Enabled:   enabled,
	}
	return request
}

func MapConfigurationRssRequest(guildID, channelID, webhookID, authorID string,
	feed entities.FeedType, enabled bool, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_CONFIGURATION_SET_RSS_REQUEST, lg)

	request.ConfigurationSetRssRequest = &amqp.ConfigurationSetRssRequest{
		GuildId:   guildID,
		ChannelId: channelID,
		FeedId:    feed.ID,
		Enabled:   enabled,
	}
	return request
}

func MapConfigurationTwitterRequest(guildID, channelID, webhookID, authorID string,
	twitterAccount entities.TwitterAccount, enabled bool, lg discordgo.Locale,
) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_CONFIGURATION_SET_TWITTER_REQUEST, lg)

	request.ConfigurationSetTwitterRequest = &amqp.ConfigurationSetTwitterRequest{
		GuildId:   guildID,
		ChannelId: channelID,
		TwitterId: twitterAccount.ID,
		Enabled:   enabled,
	}
	return request
}

func MapConfigToEmbed(guild constants.GuildConfig, emojiService emojis.Service,
	serverService servers.Service, feedService feeds.Service,
	twitterService twitters.Service, locale amqp.Language,
) *discordgo.MessageEmbed {
	lg := constants.MapAMQPLocale(locale)

	var guildServer *i18nServer
	if len(guild.ServerID) > 0 {
		server, found := serverService.GetServer(guild.ServerID)
		if !found {
			log.Warn().Str(constants.LogEntity, guild.ServerID).
				Msgf("Cannot find server based on ID sent internally, continuing with empty server")
			server = entities.Server{ID: guild.ServerID}
		}

		guildServer = &i18nServer{
			Name:  translators.GetEntityLabel(server, lg),
			Emoji: server.Emoji,
		}
	}

	serverChannels := make([]i18nChannelServer, 0)
	for _, serverChannel := range guild.ServerChannels {
		server, found := serverService.GetServer(serverChannel.ServerID)
		if !found {
			log.Warn().Str(constants.LogEntity, serverChannel.ServerID).
				Msgf("Cannot find server based on ID sent internally, continuing with empty server")
			server = entities.Server{ID: serverChannel.ServerID}
		}

		serverChannels = append(serverChannels, i18nChannelServer{
			Channel: serverChannel.Channel.Mention(),
			Server: i18nServer{
				Name:  translators.GetEntityLabel(server, lg),
				Emoji: server.Emoji,
			},
		})
	}

	notifiedChannels := mapNotifiedChannelsToI18n(guild.NotifiedChannels,
		emojiService, feedService, twitterService, lg)

	return &discordgo.MessageEmbed{
		Title: guild.Name,
		Description: i18n.Get(lg, "config.embed.description", i18n.Vars{
			"server": guildServer,
			"game":   constants.GetGame(),
		}),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: guild.Icon},
		Color:     constants.Color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: i18n.Get(lg, "config.embed.server.name", i18n.Vars{
					"gameLogo": emojiService.GetMiscStringEmoji(constants.EmojiIDGame),
				}),
				Value:  i18n.Get(lg, "config.embed.server.value", i18n.Vars{"channels": serverChannels}),
				Inline: false,
			},
			{
				Name:   i18n.Get(lg, "config.embed.webhook.name"),
				Value:  i18n.Get(lg, "config.embed.webhook.value", i18n.Vars{"channels": notifiedChannels}),
				Inline: false,
			},
		},
		Footer: discord.BuildDefaultFooter(lg),
	}
}

/** TODO integrate feed / twitter icons/title
func mapRssWebhooksToI18n(webhooks []constants.RssWebhook, emojiService emojis.Service,
	feedService feeds.Service, lg discordgo.Locale) []i18nChannelWebhook {
	i18nWebhooks := make([]i18nChannelWebhook, 0)
	for _, webhook := range webhooks {
		var providerName string
		feeds := feedService.FindFeedTypes(webhook.FeedID, lg, constants.MaxChoices)
		if len(feeds) == 1 {
			providerName = translators.GetEntityLabel(feeds[0], lg)
		} else {
			log.Warn().Str(constants.LogEntity, webhook.FeedID).
				Msgf("Cannot find feed type based on ID sent internally, continuing with default feed label")
			providerName = i18n.Get(lg, "webhooks.RSS.name")
		}

		i18nWebhooks = append(i18nWebhooks, i18nChannelWebhook{
			Channel: webhook.Channel.Mention(),
			Provider: i18nProvider{
				Name:  providerName,
				Emoji: emojiService.GetMiscStringEmoji(constants.EmojiIDRSS),
			},
		})
	}
	return i18nWebhooks
}

func mapTwitterWebhooksToI18n(webhooks []constants.TwitterWebhook, emojiService emojis.Service,
	twitterService twitters.Service, lg discordgo.Locale) []i18nChannelWebhook {
	i18nWebhooks := make([]i18nChannelWebhook, 0)
	for _, webhook := range webhooks {
		var providerName string
		twitterAccount := twitterService.GetTwitterAccount(webhook.TwitterID)
		if twitterAccount != nil {
			providerName = translators.GetEntityLabel(twitterAccount, lg)
		} else {
			log.Warn().Str(constants.LogEntity, webhook.TwitterID).
				Msgf("Cannot find twitter account based on ID sent internally, ignoring this webhook")
			continue
		}

		i18nWebhooks = append(i18nWebhooks, i18nChannelWebhook{
			Channel: webhook.Channel.Mention(),
			Provider: i18nProvider{
				Name:  providerName,
				Emoji: emojiService.GetMiscStringEmoji(constants.EmojiIDTwitter),
			},
		})
	}
	return i18nWebhooks
}
**/

func mapNotifiedChannelsToI18n(webhooks []constants.NotifiedChannel, emojiService emojis.Service,
	feedService feeds.Service, twitterService twitters.Service, lg discordgo.Locale) []i18nChannelWebhook {
	i18nWebhooks := make([]i18nChannelWebhook, 0)
	for _, webhook := range webhooks {
		i18nWebhooks = append(i18nWebhooks, i18nChannelWebhook{
			Channel: webhook.Channel.Mention(),
			Provider: i18nProvider{
				Name:  i18n.Get(lg, "webhooks.ALMANAX.name"),
				Emoji: emojiService.GetMiscStringEmoji(constants.EmojiIDAlmanax),
			},
		})
	}
	return i18nWebhooks
}
