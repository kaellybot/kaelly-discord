package mappers

import (
	"sort"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

type i18nJobExperience struct {
	Job   string
	Level int64
}

func MapBookJobGetBookRequest(jobId, serverId string, userIds []string,
	craftsmenListLimit int64, lg discordgo.Locale) *amqp.RabbitMQMessage {

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_BOOK_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		JobGetBookRequest: &amqp.JobGetBookRequest{
			UserIds:  userIds,
			JobId:    jobId,
			ServerId: serverId,
			Limit:    craftsmenListLimit,
		},
	}
}

func MapBookJobGetUserRequest(userId, serverId string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_USER_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		JobGetUserRequest: &amqp.JobGetUserRequest{
			UserId:   userId,
			ServerId: serverId,
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

func MapJobBookToEmbed(craftsmen []constants.JobUserLevel, jobId, serverId string, jobService books.BookService,
	serverService servers.ServerService, locale amqp.Language) *[]*discordgo.MessageEmbed {

	lg := constants.MapAmqpLocale(locale)

	job, found := jobService.GetJob(jobId)
	if !found {
		log.Warn().Str(constants.LogEntity, jobId).
			Msgf("Cannot find job based on ID sent internally, continuing with empty job")
		job = entities.Job{Id: jobId}
	}

	server, found := serverService.GetServer(serverId)
	if !found {
		log.Warn().Str(constants.LogEntity, serverId).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{Id: serverId}
	}

	sort.SliceStable(craftsmen, func(i, j int) bool {
		if craftsmen[i].Level == craftsmen[j].Level {
			return craftsmen[i].Username < craftsmen[j].Username
		}
		return craftsmen[i].Level > craftsmen[j].Level
	})

	embed := discordgo.MessageEmbed{
		Title:       i18n.Get(lg, "job.embed.craftsmen.title", i18n.Vars{"job": translators.GetEntityLabel(job, lg)}),
		Description: i18n.Get(lg, "job.embed.craftsmen.description", i18n.Vars{"craftsmen": craftsmen}),
		Color:       job.Color,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: job.Icon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	return &[]*discordgo.MessageEmbed{&embed}
}

func MapJobUserToEmbed(jobExperiences []*amqp.JobGetUserAnswer_JobExperience, member *discordgo.Member,
	serverId string, jobService books.BookService, serverService servers.ServerService,
	locale amqp.Language) *[]*discordgo.MessageEmbed {

	lg := constants.MapAmqpLocale(locale)
	server, found := serverService.GetServer(serverId)
	if !found {
		log.Warn().Str(constants.LogEntity, serverId).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{Id: serverId}
	}

	i18nJobXp := make([]i18nJobExperience, 0)
	for _, jobXp := range jobExperiences {
		job, found := jobService.GetJob(jobXp.JobId)
		if !found {
			log.Warn().Str(constants.LogEntity, jobXp.JobId).
				Msgf("Cannot find job based on ID sent internally, continuing with empty job")
			job = entities.Job{Id: jobXp.JobId}
		}

		i18nJobXp = append(i18nJobXp, i18nJobExperience{
			Job:   translators.GetEntityLabel(job, lg),
			Level: jobXp.Level,
		})
	}

	sort.SliceStable(i18nJobXp, func(i, j int) bool {
		if i18nJobXp[i].Level == i18nJobXp[j].Level {
			return i18nJobXp[i].Job < i18nJobXp[j].Job
		}
		return i18nJobXp[i].Level > i18nJobXp[j].Level
	})

	userName := member.Nick
	if len(userName) == 0 {
		userName = member.User.Username
	}

	userIcon := member.AvatarURL("")
	if len(userIcon) == 0 {
		userIcon = member.User.AvatarURL("")
	}

	embed := discordgo.MessageEmbed{
		Title:       i18n.Get(lg, "job.embed.craftsman.title", i18n.Vars{"username": userName}),
		Description: i18n.Get(lg, "job.embed.craftsman.description", i18n.Vars{"jobs": i18nJobXp}),
		Color:       constants.Color,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: userIcon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	return &[]*discordgo.MessageEmbed{&embed}
}
