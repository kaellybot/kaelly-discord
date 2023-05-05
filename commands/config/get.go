package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/rs/zerolog/log"
)

func (command *Command) getRequest(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {
	msg := mappers.MapConfigurationGetRequest(i.Interaction.GuildID, lg)
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
				mappers.MapConfigToEmbed(guild, command.serverService, command.feedService, message.Language),
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

	result := constants.GuildConfig{
		Name:     guild.Name,
		Icon:     guild.IconURL(),
		ServerID: answer.ServerId,
	}

	channels := make(map[string]*discordgo.Channel)
	for _, channelServer := range answer.ChannelServers {
		channel, found := channels[channelServer.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(channelServer.ChannelId)
			if errChan != nil {
				log.Warn().Err(err).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, channelServer.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			channels[channelServer.ChannelId] = discordChannel
			channel = discordChannel
		}

		result.ChannelServers = append(result.ChannelServers, constants.ChannelServer{
			Channel:  channel,
			ServerID: channelServer.ServerId,
		})
	}

	for _, webhook := range answer.AlmanaxWebhooks {
		channel, found := channels[webhook.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(webhook.ChannelId)
			if errChan != nil {
				log.Warn().Err(err).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, webhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			channels[webhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		if webhookExists(s, webhook.WebhookId, webhook.ChannelId, answer.GuildId) {
			result.AlmanaxWebhooks = append(result.AlmanaxWebhooks, constants.AlmanaxWebhook{
				ChannelWebhook: constants.ChannelWebhook{
					Channel: channel,
					Locale:  webhook.Language,
				},
			})
		}
	}

	for _, webhook := range answer.RssWebhooks {
		channel, found := channels[webhook.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(webhook.ChannelId)
			if errChan != nil {
				log.Warn().Err(err).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, webhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			channels[webhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		if webhookExists(s, webhook.WebhookId, webhook.ChannelId, answer.GuildId) {
			result.RssWebhooks = append(result.RssWebhooks, constants.RssWebhook{
				ChannelWebhook: constants.ChannelWebhook{
					Channel: channel,
					Locale:  webhook.Language,
				},
				FeedID: webhook.FeedId,
			})
		}
	}

	for _, webhook := range answer.TwitterWebhooks {
		channel, found := channels[webhook.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(webhook.ChannelId)
			if errChan != nil {
				log.Warn().Err(err).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, webhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			channels[webhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		if webhookExists(s, webhook.WebhookId, webhook.ChannelId, answer.GuildId) {
			result.TwitterWebhooks = append(result.TwitterWebhooks, constants.TwitterWebhook{
				ChannelWebhook: constants.ChannelWebhook{
					Channel: channel,
					Locale:  webhook.Language,
				},
				TwitterName: webhook.Name,
			})
		}
	}

	return result, nil
}

func webhookExists(s *discordgo.Session, webhookID, channelID, guildID string) bool {
	webhook, err := s.Webhook(webhookID)
	if err != nil {
		return false
	}

	return webhook.ChannelID == channelID && webhook.GuildID == guildID
}
