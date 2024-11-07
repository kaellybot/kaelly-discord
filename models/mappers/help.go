package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
)

func MapHelpRequest(authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return requestBackbone(authorID, amqp.RabbitMQMessage_HELP_REQUEST, lg)
}
