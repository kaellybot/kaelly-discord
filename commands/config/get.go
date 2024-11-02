package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/rs/zerolog/log"
)

func (command *Command) getRequest(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	msg := mappers.MapConfigurationGetRequest(i.Interaction.GuildID, i.Locale)
	err := command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.getRespond)
	if err != nil {
		panic(err)
	}
}

func (command *Command) getRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if message.Status == amqp.RabbitMQMessage_SUCCESS {
		guild, err := command.getGuildConfigData(s, message.ConfigurationGetAnswer)
		if err != nil {
			panic(err)
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				mappers.MapConfigToEmbed(guild, command.emojiService, command.serverService,
					command.feedService, command.videastService, command.twitterService,
					command.streamerService, message.Language),
			},
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}
	} else {
		panic(commands.ErrInvalidAnswerMessage)
	}
}

func (command *Command) getGuildConfigData(s *discordgo.Session,
	answer *amqp.ConfigurationGetAnswer) (constants.GuildConfig, error) {
	guild, err := s.Guild(answer.GuildId)
	if err != nil {
		return constants.GuildConfig{}, err
	}

	cache := make(map[string]*discordgo.Channel)
	result := constants.GuildConfig{
		Name:            guild.Name,
		Icon:            guild.IconURL(defaultIconSize),
		ServerID:        answer.ServerId,
		ChannelServers:  getValidChannelServers(s, answer, cache),
		AlmanaxWebhooks: getValidAlmanaxWebhooks(s, answer, cache),
		RssWebhooks:     getValidRSSWebhooks(s, answer, cache),
		TwitchWebhooks:  getValidTwitchWebhooks(s, answer, cache),
		TwitterWebhooks: getValidTwitterWebhooks(s, answer, cache),
		YoutubeWebhooks: getValidYoutubeWebhooks(s, answer, cache),
	}

	return result, nil
}

func getValidChannelServers(s *discordgo.Session, answer *amqp.ConfigurationGetAnswer,
	cache map[string]*discordgo.Channel) []constants.ChannelServer {
	result := make([]constants.ChannelServer, 0)
	for _, channelServer := range answer.ChannelServers {
		channel, found := cache[channelServer.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(channelServer.ChannelId)
			if errChan != nil {
				log.Warn().Err(errChan).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, channelServer.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			cache[channelServer.ChannelId] = discordChannel
			channel = discordChannel
		}

		result = append(result, constants.ChannelServer{
			Channel:  channel,
			ServerID: channelServer.ServerId,
		})
	}

	return result
}

//nolint:dupl // the code is duplicate but quite difficult to refactor: the needs behind are not the same
func getValidAlmanaxWebhooks(s *discordgo.Session, answer *amqp.ConfigurationGetAnswer,
	cache map[string]*discordgo.Channel) []constants.AlmanaxWebhook {
	result := make([]constants.AlmanaxWebhook, 0)
	for _, webhook := range answer.AlmanaxWebhooks {
		channel, found := cache[webhook.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(webhook.ChannelId)
			if errChan != nil {
				log.Warn().Err(errChan).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, webhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			cache[webhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		if webhookExists(s, webhook.WebhookId, webhook.ChannelId, answer.GuildId) {
			result = append(result, constants.AlmanaxWebhook{
				ChannelWebhook: constants.ChannelWebhook{
					Channel: channel,
					Locale:  webhook.Language,
				},
			})
		}
	}

	return result
}

func getValidRSSWebhooks(s *discordgo.Session, answer *amqp.ConfigurationGetAnswer,
	cache map[string]*discordgo.Channel) []constants.RssWebhook {
	result := make([]constants.RssWebhook, 0)
	for _, webhook := range answer.RssWebhooks {
		channel, found := cache[webhook.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(webhook.ChannelId)
			if errChan != nil {
				log.Warn().Err(errChan).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, webhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			cache[webhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		if webhookExists(s, webhook.WebhookId, webhook.ChannelId, answer.GuildId) {
			result = append(result, constants.RssWebhook{
				ChannelWebhook: constants.ChannelWebhook{
					Channel: channel,
					Locale:  webhook.Language,
				},
				FeedID: webhook.FeedId,
			})
		}
	}

	return result
}

//nolint:dupl // the code is duplicate but quite difficult to refactor: the needs behind are not the same
func getValidTwitchWebhooks(s *discordgo.Session, answer *amqp.ConfigurationGetAnswer,
	cache map[string]*discordgo.Channel) []constants.TwitchWebhook {
	result := make([]constants.TwitchWebhook, 0)
	for _, webhook := range answer.TwitchWebhooks {
		channel, found := cache[webhook.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(webhook.ChannelId)
			if errChan != nil {
				log.Warn().Err(errChan).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, webhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			cache[webhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		if webhookExists(s, webhook.WebhookId, webhook.ChannelId, answer.GuildId) {
			result = append(result, constants.TwitchWebhook{
				ChannelWebhook: constants.ChannelWebhook{
					Channel: channel,
				},
				StreamerID: webhook.StreamerId,
			})
		}
	}

	return result
}

func getValidTwitterWebhooks(s *discordgo.Session, answer *amqp.ConfigurationGetAnswer,
	cache map[string]*discordgo.Channel) []constants.TwitterWebhook {
	result := make([]constants.TwitterWebhook, 0)
	for _, webhook := range answer.TwitterWebhooks {
		channel, found := cache[webhook.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(webhook.ChannelId)
			if errChan != nil {
				log.Warn().Err(errChan).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, webhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			cache[webhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		if webhookExists(s, webhook.WebhookId, webhook.ChannelId, answer.GuildId) {
			result = append(result, constants.TwitterWebhook{
				ChannelWebhook: constants.ChannelWebhook{
					Channel: channel,
				},
				TwitterID: webhook.TwitterId,
			})
		}
	}

	return result
}

//nolint:dupl // the code is duplicate but quite difficult to refactor: the needs behind are not the same
func getValidYoutubeWebhooks(s *discordgo.Session, answer *amqp.ConfigurationGetAnswer,
	cache map[string]*discordgo.Channel) []constants.YoutubeWebhook {
	result := make([]constants.YoutubeWebhook, 0)
	for _, webhook := range answer.YoutubeWebhooks {
		channel, found := cache[webhook.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(webhook.ChannelId)
			if errChan != nil {
				log.Warn().Err(errChan).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, webhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			cache[webhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		if webhookExists(s, webhook.WebhookId, webhook.ChannelId, answer.GuildId) {
			result = append(result, constants.YoutubeWebhook{
				ChannelWebhook: constants.ChannelWebhook{
					Channel: channel,
				},
				VideastID: webhook.VideastId,
			})
		}
	}

	return result
}

func webhookExists(s *discordgo.Session, webhookID, channelID, guildID string) bool {
	webhook, err := s.Webhook(webhookID)
	if err != nil {
		return false
	}

	return webhook.ChannelID == channelID && webhook.GuildID == guildID
}
