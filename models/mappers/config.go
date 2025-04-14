package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/twitters"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	di18n "github.com/kaysoro/discordgo-i18n"
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

func MapConfigurationNotificationRequest(guildID, channelID, webhookID, authorID, labelID string,
	notificationType amqp.NotificationType, enabled bool, lg discordgo.Locale,
) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_CONFIGURATION_SET_NOTIFICATION_REQUEST, lg)
	request.ConfigurationSetNotificationRequest = &amqp.ConfigurationSetNotificationRequest{
		GuildId:          guildID,
		ChannelId:        channelID,
		WebhookId:        webhookID,
		Label:            labelID,
		NotificationType: notificationType,
		Enabled:          enabled,
	}
	return request
}

func MapConfigToEmbed(guild constants.GuildConfig, emojiService emojis.Service,
	serverService servers.Service, feedService feeds.Service,
	twitterService twitters.Service, locale amqp.Language,
) *discordgo.MessageEmbed {
	lg := i18n.MapAMQPLocale(locale)

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
			Emoji: emojiService.GetEntityStringEmoji(server.ID, constants.EmojiTypeServer),
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
				Emoji: emojiService.GetEntityStringEmoji(server.ID, constants.EmojiTypeServer),
			},
		})
	}

	notifiedChannels := mapNotifiedChannelsToI18n(guild.NotifiedChannels,
		emojiService, feedService, twitterService, lg)

	return &discordgo.MessageEmbed{
		Title: guild.Name,
		Description: di18n.Get(lg, "config.embed.description", di18n.Vars{
			"server": guildServer,
			"game":   constants.GetGame(),
		}),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: guild.Icon},
		Color:     constants.Color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: di18n.Get(lg, "config.embed.server.name", di18n.Vars{
					"gameLogo": emojiService.GetMiscStringEmoji(constants.EmojiIDGame),
				}),
				Value:  di18n.Get(lg, "config.embed.server.value", di18n.Vars{"channels": serverChannels}),
				Inline: false,
			},
			{
				Name:   di18n.Get(lg, "config.embed.webhook.name"),
				Value:  di18n.Get(lg, "config.embed.webhook.value", di18n.Vars{"channels": notifiedChannels}),
				Inline: false,
			},
		},
		Footer: discord.BuildDefaultFooter(lg),
	}
}

func mapNotifiedChannelsToI18n(webhooks []constants.NotifiedChannel, emojiService emojis.Service,
	feedService feeds.Service, twitterService twitters.Service, lg discordgo.Locale) []i18nChannelWebhook {
	i18nWebhooks := make([]i18nChannelWebhook, 0)
	for _, webhook := range webhooks {
		var provider i18nProvider
		switch webhook.NotificationType {
		case amqp.NotificationType_ALMANAX:
			provider = i18nProvider{
				Name:  di18n.Get(lg, "config.embed.webhook.almanax"),
				Emoji: emojiService.GetMiscStringEmoji(constants.EmojiIDAlmanax),
			}

		case amqp.NotificationType_RSS:
			feedType := feedService.GetFeedType(webhook.Label)
			if feedType == nil {
				log.Warn().Str(constants.LogEntity, webhook.Label).
					Msgf("Cannot find feed type based on ID sent internally, ignoring this webhook")
				continue
			}

			provider = i18nProvider{
				Name:  translators.GetEntityLabel(feedType, lg),
				Emoji: emojiService.GetMiscStringEmoji(constants.EmojiIDRSS),
			}

		case amqp.NotificationType_TWITTER:
			twitterAccount := twitterService.GetTwitterAccount(webhook.Label)
			if twitterAccount == nil {
				log.Warn().Str(constants.LogEntity, webhook.Label).
					Msgf("Cannot find twitter account based on ID sent internally, ignoring this webhook")
				continue
			}

			provider = i18nProvider{
				Name:  translators.GetEntityLabel(twitterAccount, lg),
				Emoji: emojiService.GetMiscStringEmoji(constants.EmojiIDTwitter),
			}

		case amqp.NotificationType_UNKNOWN:
			fallthrough
		default:
			continue
		}

		i18nWebhooks = append(i18nWebhooks, i18nChannelWebhook{
			Channel:  webhook.Channel.Mention(),
			Provider: provider,
		})
	}
	return i18nWebhooks
}
