package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/rs/zerolog/log"
)

func (command *Command) getRequest(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapConfigurationGetRequest(i.Interaction.GuildID, authorID, i.Locale)
	err := command.requestManager.Request(s, i, constants.ConfigurationRequestRoutingKey,
		msg, command.getRespond)
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
					command.feedService, command.twitterService, message.Language),
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
		Name:             guild.Name,
		Icon:             guild.IconURL(defaultIconSize),
		ServerID:         answer.ServerId,
		ServerChannels:   getValidServerChannels(s, answer, cache),
		NotifiedChannels: getValidNotifiedChannels(s, answer, cache),
	}

	return result, nil
}

func getValidServerChannels(s *discordgo.Session, answer *amqp.ConfigurationGetAnswer,
	cache map[string]*discordgo.Channel) []constants.ServerChannel {
	result := make([]constants.ServerChannel, 0)
	for _, serverChannel := range answer.ServerChannels {
		channel, found := cache[serverChannel.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(serverChannel.ChannelId)
			if errChan != nil {
				log.Warn().Err(errChan).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, serverChannel.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			cache[serverChannel.ChannelId] = discordChannel
			channel = discordChannel
		}

		result = append(result, constants.ServerChannel{
			Channel:  channel,
			ServerID: serverChannel.ServerId,
		})
	}

	return result
}

func getValidNotifiedChannels(s *discordgo.Session, answer *amqp.ConfigurationGetAnswer,
	cache map[string]*discordgo.Channel) []constants.NotifiedChannel {
	result := make([]constants.NotifiedChannel, 0)
	for _, notifiedChan := range answer.NotifiedChannels {
		channel, found := cache[notifiedChan.ChannelId]
		if !found {
			discordChannel, errChan := s.Channel(notifiedChan.ChannelId)
			if errChan != nil {
				log.Warn().Err(errChan).
					Str(constants.LogGuildID, answer.GuildId).
					Str(constants.LogChannelID, notifiedChan.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			cache[notifiedChan.ChannelId] = discordChannel
			channel = discordChannel
		}

		// TODO News followed?
		if webhookExists(s, "", "", answer.GuildId) {
			result = append(result, constants.NotifiedChannel{
				Channel: channel,
				// TODO
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
