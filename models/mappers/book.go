package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
)

func MapBookJobSetRequest(userId, jobId, serverId string, level uint, lg discordgo.Locale) *amqp.RabbitMQMessage {
	// TODO
	return &amqp.RabbitMQMessage{}
}
