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

type i18nAlignmentExperience struct {
	City  string
	Order string
	Level int64
}

func MapBookAlignGetBookRequest(cityId, orderId, serverId string, userIds []string,
	craftsmenListLimit int64, lg discordgo.Locale) *amqp.RabbitMQMessage {

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ALIGN_GET_BOOK_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		AlignGetBookRequest: &amqp.AlignGetBookRequest{
			UserIds:  userIds,
			CityId:   cityId,
			OrderId:  orderId,
			ServerId: serverId,
			Limit:    craftsmenListLimit,
		},
	}
}

func MapBookAlignGetUserRequest(userId, serverId string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ALIGN_GET_USER_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		AlignGetUserRequest: &amqp.AlignGetUserRequest{
			UserId:   userId,
			ServerId: serverId,
		},
	}
}

func MapBookAlignSetRequest(userId, cityId, orderId, serverId string, level int64,
	lg discordgo.Locale) *amqp.RabbitMQMessage {

	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ALIGN_SET_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		AlignSetRequest: &amqp.AlignSetRequest{
			UserId:   userId,
			CityId:   cityId,
			OrderId:  orderId,
			ServerId: serverId,
			Level:    level,
		},
	}
}

func MapAlignBookToEmbed(believers []constants.AlignmentUserLevel, serverId string, jobService books.BookService,
	serverService servers.ServerService, locale amqp.Language) *[]*discordgo.MessageEmbed {

	lg := constants.MapAmqpLocale(locale)
	server, found := serverService.GetServer(serverId)
	if !found {
		log.Warn().Str(constants.LogEntity, serverId).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{Id: serverId}
	}

	sort.SliceStable(believers, func(i, j int) bool {
		if believers[i].Level == believers[j].Level {
			return believers[i].Username < believers[j].Username
		}
		return believers[i].Level > believers[j].Level
	})

	embed := discordgo.MessageEmbed{
		Title:       i18n.Get(lg, "align.embed.believers.title"),
		Description: i18n.Get(lg, "align.embed.believers.description", i18n.Vars{"believers": believers}),
		Color:       0, // TODO EMBED color & thumbnail
		Thumbnail:   &discordgo.MessageEmbedThumbnail{},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	return &[]*discordgo.MessageEmbed{&embed}
}

func MapAlignUserToEmbed(alignExperiences []*amqp.AlignGetUserAnswer_AlignExperience, member *discordgo.Member,
	serverId string, alignService books.BookService, serverService servers.ServerService,
	locale amqp.Language) *[]*discordgo.MessageEmbed {

	lg := constants.MapAmqpLocale(locale)
	server, found := serverService.GetServer(serverId)
	if !found {
		log.Warn().Str(constants.LogEntity, serverId).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{Id: serverId}
	}

	i18nAlignXp := make([]i18nAlignmentExperience, 0)
	for _, alignXp := range alignExperiences {
		city, found := alignService.GetCity(alignXp.CityId)
		if !found {
			log.Warn().Str(constants.LogEntity, alignXp.CityId).
				Msgf("Cannot find city based on ID sent internally, continuing with empty city")
			city = entities.City{Id: alignXp.CityId}
		}

		order, found := alignService.GetOrder(alignXp.OrderId)
		if !found {
			log.Warn().Str(constants.LogEntity, alignXp.OrderId).
				Msgf("Cannot find order based on ID sent internally, continuing with empty order")
			order = entities.Order{Id: alignXp.OrderId}
		}

		i18nAlignXp = append(i18nAlignXp, i18nAlignmentExperience{
			City:  translators.GetEntityLabel(city, lg),
			Order: translators.GetEntityLabel(order, lg),
			Level: alignXp.Level,
		})
	}

	sort.SliceStable(i18nAlignXp, func(i, j int) bool {
		if i18nAlignXp[i].Level == i18nAlignXp[j].Level {
			if i18nAlignXp[i].City == i18nAlignXp[j].City {
				return i18nAlignXp[i].Order < i18nAlignXp[j].Order
			}
			return i18nAlignXp[i].City < i18nAlignXp[j].City
		}
		return i18nAlignXp[i].Level > i18nAlignXp[j].Level
	})

	userName := member.Nick
	if len(userName) == 0 {
		userName = member.User.Username
	}

	userIcon := member.AvatarURL("")
	if len(userIcon) == 0 {
		userIcon = member.User.AvatarURL("")
	}

	// TODO EMBED

	embed := discordgo.MessageEmbed{
		Title:       i18n.Get(lg, "align.embed.beliefs.title", i18n.Vars{"username": userName}),
		Description: i18n.Get(lg, "align.embed.beliefs.description", i18n.Vars{"beliefs": i18nAlignXp}),
		Color:       constants.Color,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: userIcon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	return &[]*discordgo.MessageEmbed{&embed}
}
