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
	Username string
	City     string
	Order    string
	Level    int64
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

func MapAlignBookToEmbed(believers []constants.AlignmentUserLevel, serverId string, alignService books.BookService,
	serverService servers.ServerService, locale amqp.Language) *[]*discordgo.MessageEmbed {

	lg := constants.MapAmqpLocale(locale)
	server, found := serverService.GetServer(serverId)
	if !found {
		log.Warn().Str(constants.LogEntity, serverId).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{Id: serverId}
	}

	cityValues := make(map[string]int64)
	i18nAlignXp := make([]i18nAlignmentExperience, 0)
	for _, alignXp := range believers {
		value, found := cityValues[alignXp.CityId]
		if found {
			cityValues[alignXp.CityId] = value + alignXp.Level
		} else {
			cityValues[alignXp.CityId] = alignXp.Level
		}

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
			Username: alignXp.Username,
			City:     city.Emoji,
			Order:    order.Emoji,
			Level:    alignXp.Level,
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

	winningCity := mapWinningCity(cityValues, alignService)
	embed := discordgo.MessageEmbed{
		Title:       i18n.Get(lg, "align.embed.believers.title"),
		Description: i18n.Get(lg, "align.embed.believers.description", i18n.Vars{"believers": i18nAlignXp}),
		Color:       winningCity.Color,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: winningCity.Icon},
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

	cityValues := make(map[string]int64)
	i18nAlignXp := make([]i18nAlignmentExperience, 0)
	for _, alignXp := range alignExperiences {
		value, found := cityValues[alignXp.CityId]
		if found {
			cityValues[alignXp.CityId] = value + alignXp.Level
		} else {
			cityValues[alignXp.CityId] = alignXp.Level
		}

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
			City:  city.Emoji,
			Order: order.Emoji,
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

	winningCity := mapWinningCity(cityValues, alignService)
	embed := discordgo.MessageEmbed{
		Title:       i18n.Get(lg, "align.embed.beliefs.title", i18n.Vars{"username": userName}),
		Description: i18n.Get(lg, "align.embed.beliefs.description", i18n.Vars{"beliefs": i18nAlignXp}),
		Color:       winningCity.Color,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: userIcon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	return &[]*discordgo.MessageEmbed{&embed}
}

func mapWinningCity(cityValues map[string]int64, alignService books.BookService) entities.City {
	var winningCity = constants.NeutralCity
	var winningValue int64 = 0
	for cityId, value := range cityValues {
		if winningValue == value {
			winningCity = constants.NeutralCity
		} else if winningValue < value {
			winningValue = value
			city, found := alignService.GetCity(cityId)
			if !found {
				log.Warn().Str(constants.LogEntity, cityId).
					Msgf("Cannot find city based on ID sent internally, continuing with neutral city")
				city = constants.NeutralCity
			}

			winningCity = city
		}
	}

	return winningCity
}
