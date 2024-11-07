package mappers

import (
	"sort"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

type i18nJobExperience struct {
	Job   string
	Level int64
}

func MapBookJobGetBookRequest(jobID, serverID string, page int,
	userIDs []string, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_BOOK_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		Game:     constants.GetGame().AMQPGame,
		JobGetBookRequest: &amqp.JobGetBookRequest{
			UserIds:  userIDs,
			JobId:    jobID,
			ServerId: serverID,
			Offset:   int32(page) * constants.MaxBookRowPerEmbed,
			Size:     constants.MaxBookRowPerEmbed,
		},
	}
}

func MapBookJobGetUserRequest(userID, serverID, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_USER_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		Game:     constants.GetGame().AMQPGame,
		JobGetUserRequest: &amqp.JobGetUserRequest{
			UserId:   userID,
			ServerId: serverID,
		},
	}
}

func MapBookJobSetRequest(userID, jobID, serverID string, level int64,
	lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(userID, amqp.RabbitMQMessage_COMPETITION_MAP_REQUEST, lg)
	request.JobSetRequest = &amqp.JobSetRequest{
		UserId:   userID,
		JobId:    jobID,
		ServerId: serverID,
		Level:    level,
	}
	return request
}

func MapJobBookToWebhook(answer *amqp.JobGetBookAnswer,
	craftsmen []constants.JobUserLevel, jobService books.Service,
	serverService servers.Service, emojiService emojis.Service,
	lg discordgo.Locale) *discordgo.WebhookEdit {
	return &discordgo.WebhookEdit{
		Embeds:     mapJobBookToEmbeds(answer, craftsmen, jobService, serverService, lg),
		Components: mapJobBookToComponents(answer, lg, emojiService),
	}
}

func mapJobBookToComponents(answer *amqp.JobGetBookAnswer,
	lg discordgo.Locale, emojiService emojis.Service) *[]discordgo.MessageComponent {
	jobID := answer.GetJobId()
	serverID := answer.GetServerId()
	crafter := func(page int) string {
		return contract.CraftJobBookCustomID(jobID, serverID, page)
	}

	paginations := discord.GetPaginationButtons(int(answer.GetPage()), int(answer.GetPages()),
		crafter, lg, emojiService)
	if len(paginations) == 0 {
		return &paginations
	}

	return &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: paginations,
		},
	}
}

func mapJobBookToEmbeds(answer *amqp.JobGetBookAnswer, craftsmen []constants.JobUserLevel, jobService books.Service,
	serverService servers.Service, lg discordgo.Locale) *[]*discordgo.MessageEmbed {
	job, found := jobService.GetJob(answer.GetJobId())
	if !found {
		log.Warn().Str(constants.LogEntity, answer.GetJobId()).
			Msgf("Cannot find job based on ID sent internally, continuing with empty job")
		job = entities.Job{ID: answer.GetJobId()}
	}

	server, found := serverService.GetServer(answer.GetServerId())
	if !found {
		log.Warn().Str(constants.LogEntity, answer.GetServerId()).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{ID: answer.GetServerId()}
	}

	sort.SliceStable(craftsmen, func(i, j int) bool {
		if craftsmen[i].Level == craftsmen[j].Level {
			return craftsmen[i].Username < craftsmen[j].Username
		}
		return craftsmen[i].Level > craftsmen[j].Level
	})

	embed := discordgo.MessageEmbed{
		Title: i18n.Get(lg, "job.embed.craftsmen.title", i18n.Vars{"job": translators.GetEntityLabel(job, lg)}),
		Description: i18n.Get(lg, "job.embed.craftsmen.description", i18n.Vars{
			"craftsmen": craftsmen,
			"total":     answer.GetTotal(),
			"page":      answer.GetPage() + 1,
			"pages":     answer.GetPages(),
		}),
		Color:     job.Color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: job.Icon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	return &[]*discordgo.MessageEmbed{&embed}
}

func MapJobUserToEmbed(jobExperiences []*amqp.JobGetUserAnswer_JobExperience, member *discordgo.Member,
	serverID string, jobService books.Service, serverService servers.Service,
	locale amqp.Language) *[]*discordgo.MessageEmbed {
	lg := constants.MapAMQPLocale(locale)
	server, found := serverService.GetServer(serverID)
	if !found {
		log.Warn().Str(constants.LogEntity, serverID).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{ID: serverID}
	}

	i18nJobXp := make([]i18nJobExperience, 0)
	for _, jobXp := range jobExperiences {
		job, foundJob := jobService.GetJob(jobXp.JobId)
		if !foundJob {
			log.Warn().Str(constants.LogEntity, jobXp.JobId).
				Msgf("Cannot find job based on ID sent internally, continuing with empty job")
			job = entities.Job{ID: jobXp.JobId}
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
