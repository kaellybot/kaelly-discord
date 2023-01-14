package mappers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
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

func MapJobBookToEmbed(jobBook *amqp.JobGetAnswer, jobService books.BookService,
	serverService servers.ServerService, locale amqp.RabbitMQMessage_Language) *[]*discordgo.MessageEmbed {

	lg := constants.MapAmqpLocale(locale)

	job, found := jobService.GetJob(jobBook.JobId)
	if !found {
		log.Warn().Str(constants.LogEntity, jobBook.JobId).
			Msgf("Cannot find job based on ID sent internally, continuing with empty job")
		job = entities.Job{Id: jobBook.JobId}
	}

	server, found := serverService.GetServer(jobBook.ServerId)
	if !found {
		log.Warn().Str(constants.LogEntity, jobBook.ServerId).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{Id: jobBook.ServerId}
	}

	embed := discordgo.MessageEmbed{
		Title:       translators.GetEntityLabel(job, lg),
		Description: fmt.Sprintf("%v", jobBook.Craftsmen), // TODO members
		Color:       job.Color,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: job.Icon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	return &[]*discordgo.MessageEmbed{&embed}
}
