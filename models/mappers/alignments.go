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

func MapBookAlignGetBookRequest(cityID, orderID, serverID string, userIDs []string,
	craftsmenListLimit int64, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ALIGN_GET_BOOK_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		AlignGetBookRequest: &amqp.AlignGetBookRequest{
			UserIds:  userIDs,
			CityId:   cityID,
			OrderId:  orderID,
			ServerId: serverID,
			Limit:    craftsmenListLimit,
		},
	}
}

func MapBookAlignGetUserRequest(userID, serverID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ALIGN_GET_USER_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		AlignGetUserRequest: &amqp.AlignGetUserRequest{
			UserId:   userID,
			ServerId: serverID,
		},
	}
}

func MapBookAlignSetRequest(userID, cityID, orderID, serverID string, level int64,
	lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ALIGN_SET_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		AlignSetRequest: &amqp.AlignSetRequest{
			UserId:   userID,
			CityId:   cityID,
			OrderId:  orderID,
			ServerId: serverID,
			Level:    level,
		},
	}
}

func MapAlignBookToEmbed(believers []constants.AlignmentUserLevel, serverID string, alignService books.Service,
	serverService servers.Service, locale amqp.Language) *[]*discordgo.MessageEmbed {
	lg := constants.MapAMQPLocale(locale)
	server, found := serverService.GetServer(serverID)
	if !found {
		log.Warn().Str(constants.LogEntity, serverID).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{ID: serverID}
	}

	cityValues := make(map[string]int64)
	i18nAlignXp := make([]i18nAlignmentExperience, 0)
	for _, alignXp := range believers {
		value, foundAlign := cityValues[alignXp.CityID]
		if foundAlign {
			cityValues[alignXp.CityID] = value + alignXp.Level
		} else {
			cityValues[alignXp.CityID] = alignXp.Level
		}

		city, foundCity := alignService.GetCity(alignXp.CityID)
		if !foundCity {
			log.Warn().Str(constants.LogEntity, alignXp.CityID).
				Msgf("Cannot find city based on ID sent internally, continuing with empty city")
			city = entities.City{ID: alignXp.CityID}
		}

		order, foundOrder := alignService.GetOrder(alignXp.OrderID)
		if !foundOrder {
			log.Warn().Str(constants.LogEntity, alignXp.OrderID).
				Msgf("Cannot find order based on ID sent internally, continuing with empty order")
			order = entities.Order{ID: alignXp.OrderID}
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
	serverID string, alignService books.Service, serverService servers.Service,
	locale amqp.Language) *[]*discordgo.MessageEmbed {
	lg := constants.MapAMQPLocale(locale)
	server, found := serverService.GetServer(serverID)
	if !found {
		log.Warn().Str(constants.LogEntity, serverID).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{ID: serverID}
	}

	cityValues := make(map[string]int64)
	i18nAlignXp := make([]i18nAlignmentExperience, 0)
	for _, alignXp := range alignExperiences {
		value, foundCity := cityValues[alignXp.CityId]
		if foundCity {
			cityValues[alignXp.CityId] = value + alignXp.Level
		} else {
			cityValues[alignXp.CityId] = alignXp.Level
		}

		city, foundCity := alignService.GetCity(alignXp.CityId)
		if !foundCity {
			log.Warn().Str(constants.LogEntity, alignXp.CityId).
				Msgf("Cannot find city based on ID sent internally, continuing with empty city")
			city = entities.City{ID: alignXp.CityId}
		}

		order, foundOrder := alignService.GetOrder(alignXp.OrderId)
		if !foundOrder {
			log.Warn().Str(constants.LogEntity, alignXp.OrderId).
				Msgf("Cannot find order based on ID sent internally, continuing with empty order")
			order = entities.Order{ID: alignXp.OrderId}
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

func mapWinningCity(cityValues map[string]int64, alignService books.Service) entities.City {
	var winningCity = constants.GetNeutralCity()
	var winningValue int64 = 0
	for cityID, value := range cityValues {
		if winningValue == value {
			winningCity = constants.GetNeutralCity()
		} else if winningValue < value {
			winningValue = value
			city, found := alignService.GetCity(cityID)
			if !found {
				log.Warn().Str(constants.LogEntity, cityID).
					Msgf("Cannot find city based on ID sent internally, continuing with neutral city")
				city = constants.GetNeutralCity()
			}

			winningCity = city
		}
	}

	return winningCity
}
