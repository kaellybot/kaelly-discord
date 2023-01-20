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

func (command *ConfigCommand) getRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	msg := mappers.MapConfigurationGetRequest(i.Interaction.GuildID, lg)
	err := command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.getRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) getRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	if message.Status == amqp.RabbitMQMessage_SUCCESS {
		guild, err := command.getGuildConfigData(s, message.ConfigurationGetAnswer)
		if err != nil {
			panic(err)
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				mappers.MapConfigToEmbed(guild, command.serverService, message.Language),
			},
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}
	} else {
		panic(commands.ErrInvalidAnswerMessage)
	}
}

func (command *ConfigCommand) getGuildConfigData(s *discordgo.Session,
	answer *amqp.ConfigurationGetAnswer) (constants.GuildConfig, error) {

	guild, err := s.Guild(answer.GuildId)
	if err != nil {
		return constants.GuildConfig{}, err
	}

	result := constants.GuildConfig{
		Name:     guild.Name,
		Icon:     guild.IconURL(),
		ServerId: answer.ServerId,
	}

	channels := make(map[string]*discordgo.Channel)
	for _, channelServer := range answer.ChannelServers {
		channel, found := channels[channelServer.ChannelId]
		if !found {
			discordChannel, err := s.Channel(channelServer.ChannelId)
			if err != nil {
				log.Warn().Err(err).
					Str(constants.LogGuildId, answer.GuildId).
					Str(constants.LogChannelId, channelServer.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			channels[channelServer.ChannelId] = discordChannel
			channel = discordChannel
		}

		result.ChannelServers = append(result.ChannelServers, constants.ChannelServer{
			Channel:  channel,
			ServerId: channelServer.ServerId,
		})
	}

	for _, channelWebhook := range answer.ChannelWebhooks {
		channel, found := channels[channelWebhook.ChannelId]
		if !found {
			discordChannel, err := s.Channel(channelWebhook.ChannelId)
			if err != nil {
				log.Warn().Err(err).
					Str(constants.LogGuildId, answer.GuildId).
					Str(constants.LogChannelId, channelWebhook.ChannelId).
					Msgf("Cannot retrieve channel from Discord, ignoring this line...")
				continue
			}

			channels[channelWebhook.ChannelId] = discordChannel
			channel = discordChannel
		}

		result.ChannelWebhooks = append(result.ChannelWebhooks, constants.ChannelWebhook{
			Channel: channel,
			// TODO
		})
	}

	return result, nil
}
