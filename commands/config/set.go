package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) setRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isConfigSetAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	if message.ConfigurationSetAnswer.RemoveWebhook {
		err := s.WebhookDelete(message.ConfigurationSetAnswer.WebhookId)
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot remove webhook after receiving internal answer, ignoring webhook deletion")
		}
	}

	if message.Status == amqp.RabbitMQMessage_SUCCESS {
		content := i18n.Get(constants.MapAMQPLocale(message.Language), "config.success")
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}
	} else {
		panic(commands.ErrInvalidAnswerMessage)
	}
}

func isConfigSetAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Type == amqp.RabbitMQMessage_CONFIGURATION_SET_ANSWER &&
		message.ConfigurationSetAnswer != nil
}
