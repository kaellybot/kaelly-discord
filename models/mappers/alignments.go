package mappers

import (
	"fmt"
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

type i18nAlignmentExperience struct {
	Username string
	City     string
	Order    string
	Level    int64
}

func MapBookAlignGetBookRequest(cityID, orderID, serverID string, page int,
	userIDs []string, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ALIGN_GET_BOOK_REQUEST, lg)
	request.AlignGetBookRequest = &amqp.AlignGetBookRequest{
		UserIds:  userIDs,
		CityId:   cityID,
		OrderId:  orderID,
		ServerId: serverID,
		Offset:   int64(page) * constants.MaxBookRowPerEmbed,
		Size:     constants.MaxBookRowPerEmbed,
	}
	return request
}

func MapBookAlignGetUserRequest(userID, serverID string,
	authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ALIGN_GET_USER_REQUEST, lg)
	request.AlignGetUserRequest = &amqp.AlignGetUserRequest{
		UserId:   userID,
		ServerId: serverID,
	}
	return request
}

func MapBookAlignSetRequest(userID, cityID, orderID, serverID string, level int64,
	lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(userID, amqp.RabbitMQMessage_ALIGN_SET_REQUEST, lg)
	request.AlignSetRequest = &amqp.AlignSetRequest{
		UserId:   userID,
		CityId:   cityID,
		OrderId:  orderID,
		ServerId: serverID,
		Level:    level,
	}
	return request
}

func MapAlignBookToWebhook(answer *amqp.AlignGetBookAnswer,
	believers []constants.AlignmentUserLevel, alignService books.Service,
	serverService servers.Service, emojiService emojis.Service,
	lg discordgo.Locale) *discordgo.WebhookEdit {
	return &discordgo.WebhookEdit{
		Embeds:     mapAlignBookToEmbeds(answer, believers, alignService, emojiService, serverService, lg),
		Components: mapAlignBookToComponents(answer, lg, alignService, emojiService),
	}
}

func mapAlignBookToEmbeds(answer *amqp.AlignGetBookAnswer,
	believers []constants.AlignmentUserLevel, alignService books.Service,
	emojiService emojis.Service, serverService servers.Service, lg discordgo.Locale,
) *[]*discordgo.MessageEmbed {
	server, found := serverService.GetServer(answer.GetServerId())
	if !found {
		log.Warn().Str(constants.LogEntity, answer.GetServerId()).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{ID: answer.GetServerId()}
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
			city = entities.City{}
		}

		order, foundOrder := alignService.GetOrder(alignXp.OrderID)
		if !foundOrder {
			log.Warn().Str(constants.LogEntity, alignXp.OrderID).
				Msgf("Cannot find order based on ID sent internally, continuing with empty order")
			order = entities.Order{}
		}

		orderEmojiID := buildEmojiOrderID(&city, order.ID)
		i18nAlignXp = append(i18nAlignXp, i18nAlignmentExperience{
			Username: alignXp.Username,
			City:     emojiService.GetEntityStringEmoji(city.ID, constants.EmojiTypeCity),
			Order:    emojiService.GetEntityStringEmoji(orderEmojiID, constants.EmojiTypeOrder),
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
		Title: i18n.Get(lg, "align.embed.believers.title"),
		Description: i18n.Get(lg, "align.embed.believers.description", i18n.Vars{
			"believers": i18nAlignXp,
			"total":     answer.GetTotal(),
			"page":      answer.GetPage() + 1,
			"pages":     answer.GetPages(),
		}),
		Color:     winningCity.Color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: winningCity.Icon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	return &[]*discordgo.MessageEmbed{&embed}
}

func mapAlignBookToComponents(answer *amqp.AlignGetBookAnswer, lg discordgo.Locale,
	alignService books.Service, emojiService emojis.Service) *[]discordgo.MessageComponent {
	components := make([]discordgo.MessageComponent, 0)
	cityID := answer.GetCityId()
	orderID := answer.GetOrderId()
	serverID := answer.GetServerId()
	crafter := func(page int) string {
		return contract.CraftAlignBookPageCustomID(cityID, orderID, serverID, page)
	}

	paginations := discord.GetPaginationButtons(int(answer.GetPage()), int(answer.GetPages()),
		crafter, lg, emojiService)

	if len(paginations) != 0 {
		components = append(components,
			discordgo.ActionsRow{
				Components: paginations,
			},
		)
	}

	var chosenCity *entities.City
	cityChoices := make([]discordgo.SelectMenuOption, 0)
	cityChoices = append(cityChoices, discordgo.SelectMenuOption{
		Label:   i18n.Get(lg, "align.embed.believers.placeholders.cities"),
		Value:   contract.AlignAllValues,
		Default: len(cityID) == 0,
		Emoji:   emojiService.GetMiscEmoji(constants.EmojiIDGlobal),
	})
	for _, city := range alignService.GetCities() {
		cityChoices = append(cityChoices, discordgo.SelectMenuOption{
			Label:   translators.GetEntityLabel(city, lg),
			Value:   city.ID,
			Default: city.ID == cityID,
			Emoji:   emojiService.GetEntityEmoji(city.ID, constants.EmojiTypeCity),
		})

		if city.ID == cityID {
			chosenCity = &city
		}
	}

	components = append(components,
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    contract.CraftAlignBookCityCustomID(orderID, serverID),
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: i18n.Get(lg, "align.embed.believers.placeholders.city"),
					Options:     cityChoices,
				},
			},
		},
	)

	orderChoices := make([]discordgo.SelectMenuOption, 0)
	orderChoices = append(orderChoices, discordgo.SelectMenuOption{
		Label:   i18n.Get(lg, "align.embed.believers.placeholders.orders"),
		Value:   contract.AlignAllValues,
		Default: len(orderID) == 0,
		Emoji:   emojiService.GetMiscEmoji(constants.EmojiIDGlobal),
	})

	for _, order := range alignService.GetOrders() {
		emojiOrderID := buildEmojiOrderID(chosenCity, order.ID)
		orderChoices = append(orderChoices, discordgo.SelectMenuOption{
			Label:   translators.GetEntityLabel(order, lg),
			Value:   order.ID,
			Default: order.ID == orderID,
			Emoji:   emojiService.GetEntityEmoji(emojiOrderID, constants.EmojiTypeOrder),
		})
	}

	components = append(components,
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    contract.CraftAlignBookOrderCustomID(cityID, serverID),
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: i18n.Get(lg, "align.embed.believers.placeholders.order"),
					Options:     orderChoices,
				},
			},
		},
	)

	return &components
}

func MapAlignUserToEmbed(alignExperiences []*amqp.AlignGetUserAnswer_AlignExperience,
	member *discordgo.Member, serverID string, alignService books.Service,
	emojiService emojis.Service, serverService servers.Service, locale amqp.Language,
) *[]*discordgo.MessageEmbed {
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
			city = entities.City{}
		}

		order, foundOrder := alignService.GetOrder(alignXp.OrderId)
		if !foundOrder {
			log.Warn().Str(constants.LogEntity, alignXp.OrderId).
				Msgf("Cannot find order based on ID sent internally, continuing with empty order")
			order = entities.Order{}
		}

		orderEmojiID := buildEmojiOrderID(&city, order.ID)
		i18nAlignXp = append(i18nAlignXp, i18nAlignmentExperience{
			City:  emojiService.GetEntityStringEmoji(city.ID, constants.EmojiTypeCity),
			Order: emojiService.GetEntityStringEmoji(orderEmojiID, constants.EmojiTypeOrder),
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
	var winningCity = entities.GetNeutralCity()
	var winningValue int64
	for cityID, value := range cityValues {
		if winningValue == value {
			winningCity = entities.GetNeutralCity()
		} else if winningValue < value {
			winningValue = value
			city, found := alignService.GetCity(cityID)
			if !found {
				log.Warn().Str(constants.LogEntity, cityID).
					Msgf("Cannot find city based on ID sent internally, continuing with neutral city")
				city = entities.GetNeutralCity()
			}

			winningCity = city
		}
	}

	return winningCity
}

func buildEmojiOrderID(city *entities.City, orderID string) string {
	if city == nil {
		return orderID
	}

	if city.Type == constants.CityTypeLight {
		return fmt.Sprintf("%v%v", constants.CityTypeLight, orderID)
	}

	if city.Type == constants.CityTypeDark {
		return fmt.Sprintf("%v%v", constants.CityTypeDark, orderID)
	}

	return orderID
}
