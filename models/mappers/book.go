package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

func MapBookJobGetRequest(jobId, serverId string, userIds []string,
	craftsmenListLimit int64, lg discordgo.Locale) *amqp.RabbitMQMessage {

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		JobGetRequest: &amqp.JobGetRequest{
			UserIds:  userIds,
			JobId:    jobId,
			ServerId: serverId,
			Limit:    craftsmenListLimit,
		},
	}
}

func MapBookJobSetRequest(userId, jobId, serverId string, level int64,
	lg discordgo.Locale) *amqp.RabbitMQMessage {

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_SET_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		JobSetRequest: &amqp.JobSetRequest{
			UserId:   userId,
			JobId:    jobId,
			ServerId: serverId,
			Level:    level,
		},
	}
}
