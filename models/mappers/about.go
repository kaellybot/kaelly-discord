package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
)

func MapAboutRequest(authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return requestBackbone(authorID, amqp.RabbitMQMessage_ABOUT_REQUEST, lg)
}
