package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) setNotificationRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isConfigSetNotificationAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	if message.ConfigurationSetNotificationAnswer.RemoveWebhook {
		err := s.WebhookDelete(message.ConfigurationSetNotificationAnswer.WebhookId)
		if err != nil {
			apiError, ok := discord.ExtractAPIError(err)
			if !ok || apiError.Code != constants.DiscordCodeNotFound {
				log.Warn().Err(err).
					Msgf("Cannot remove webhook after receiving internal answer, ignoring webhook deletion")
			}
		}
	}

	content := i18n.Get(constants.MapAMQPLocale(message.Language), "config.success")
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func isConfigSetNotificationAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Type == amqp.RabbitMQMessage_CONFIGURATION_SET_NOTIFICATION_ANSWER &&
		message.ConfigurationSetNotificationAnswer != nil && message.Status == amqp.RabbitMQMessage_SUCCESS
}
