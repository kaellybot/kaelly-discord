package mappers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func MapSetListRequest(query string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_SET_LIST_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaSetListRequest: &amqp.EncyclopediaSetListRequest{
			Query: query,
		},
	}
}

func MapSetRequest(query string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_SET_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaSetRequest: &amqp.EncyclopediaSetRequest{
			Query: query,
		},
	}
}

func MapSetToDefaultWebhookEdit(set *amqp.EncyclopediaSetAnswer, characService characteristics.Service,
	emojiService emojis.Service, locale amqp.Language) *discordgo.WebhookEdit {
	return MapSetToWebhookEdit(set, len(set.Equipments), characService, emojiService, locale)
}

func MapSetToWebhookEdit(set *amqp.EncyclopediaSetAnswer, itemNumber int,
	characService characteristics.Service, emojiService emojis.Service,
	locale amqp.Language) *discordgo.WebhookEdit {
	lg := constants.MapAMQPLocale(locale)
	bonus := &amqp.EncyclopediaSetAnswer_Bonus{ItemNumber: 0}
	for _, currentBonus := range set.Bonuses {
		if currentBonus.ItemNumber == int64(itemNumber) {
			bonus = currentBonus
			break
		} else if bonus.ItemNumber < currentBonus.ItemNumber {
			bonus = currentBonus
		}
	}

	if bonus != nil && bonus.ItemNumber != int64(itemNumber) {
		log.Warn().
			Str(constants.LogAnkamaID, set.Id).
			Int(constants.LogItemNumber, itemNumber).
			Msgf("Set bonus with specific item numbers was not found, returning the highest one...")
	}

	return &discordgo.WebhookEdit{
		Embeds:     mapSetToEmbeds(set, bonus, characService, lg),
		Components: mapSetToComponents(set, bonus, emojiService, lg),
	}
}

func mapSetToEmbeds(set *amqp.EncyclopediaSetAnswer, bonus *amqp.EncyclopediaSetAnswer_Bonus,
	service characteristics.Service, lg discordgo.Locale) *[]*discordgo.MessageEmbed {
	fields := discord.SliceFields(set.GetEquipments(), constants.MaxEquipmentPerField,
		func(i int, items []*amqp.EncyclopediaSetAnswer_Equipment) *discordgo.MessageEmbedField {
			name := constants.InvisibleCharacter
			if i == 0 {
				name = i18n.Get(lg, "set.items.title")
			}

			return &discordgo.MessageEmbedField{
				Name: name,
				Value: i18n.Get(lg, "set.items.description", i18n.Vars{
					"items": mapSetItems(items, lg),
				}),
				Inline: true,
			}
		})

	if bonus != nil {
		bonusFields := discord.SliceFields(bonus.GetEffects(), constants.MaxCharacterPerField,
			func(i int, items []*amqp.EncyclopediaSetAnswer_Effect) *discordgo.MessageEmbedField {
				name := constants.InvisibleCharacter
				if i == 0 {
					name = i18n.Get(lg, "set.effects.title", i18n.Vars{
						"itemNumber": bonus.GetItemNumber(),
					})
				}

				return &discordgo.MessageEmbedField{
					Name: name,
					Value: i18n.Get(lg, "set.effects.description", i18n.Vars{
						"effects": mapSetEffects(items, service),
					}),
					Inline: true,
				}
			})

		fields = append(fields, bonusFields...)
	}

	return &[]*discordgo.MessageEmbed{
		{
			Title:       set.Name,
			Description: i18n.Get(lg, "set.description", i18n.Vars{"level": set.Level}),
			Color:       constants.Color,
			URL:         i18n.Get(lg, "set.url", i18n.Vars{"id": set.Id}),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/zY6C2ai.png", // TODO URL
			},
			Fields: fields,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    set.Source.Name,
				URL:     set.Source.Url,
				IconURL: set.Source.Icon,
			},
		},
	}
}

func mapSetToComponents(set *amqp.EncyclopediaSetAnswer, bonus *amqp.EncyclopediaSetAnswer_Bonus,
	service emojis.Service, lg discordgo.Locale) *[]discordgo.MessageComponent {
	components := make([]discordgo.MessageComponent, 0)

	bonuses := make([]discordgo.SelectMenuOption, 0)
	for _, currentBonus := range set.Bonuses {
		emoji := service.GetSetBonusEmoji(int(currentBonus.ItemNumber), len(set.Equipments))
		bonuses = append(bonuses, discordgo.SelectMenuOption{
			Label: i18n.Get(lg, "set.effects.option", i18n.Vars{
				"itemNumber": currentBonus.ItemNumber,
				"itemCount":  len(set.Equipments),
			}),
			Value:   fmt.Sprintf("%v", currentBonus.ItemNumber),
			Default: currentBonus.ItemNumber == bonus.ItemNumber,
			Emoji: discordgo.ComponentEmoji{
				ID: emoji.Snowflake,
			},
		})
	}

	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    fmt.Sprintf("/sets/%v/effects", set.Id),
				MenuType:    discordgo.StringSelectMenu,
				Placeholder: i18n.Get(lg, "set.effects.placeholder"),
				Options:     bonuses,
			},
		},
	})

	items := make([]discordgo.SelectMenuOption, 0)
	for _, item := range set.Equipments {
		items = append(items, discordgo.SelectMenuOption{
			Label: item.Name,
			Value: item.Id,
			Emoji: discordgo.ComponentEmoji{
				ID: service.GetEquipmentEmoji(item.Type).Snowflake,
			},
		})
	}

	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    fmt.Sprintf("/sets/%v/items", set.Id),
				MenuType:    discordgo.StringSelectMenu,
				Placeholder: i18n.Get(lg, "set.items.placeholder"),
				Options:     items,
			},
		},
	})

	return &components
}

type i18nItem struct {
	Name string
	URL  string
}

func mapSetItems(items []*amqp.EncyclopediaSetAnswer_Equipment, lg discordgo.Locale) []i18nItem {
	result := make([]i18nItem, 0)
	for _, item := range items {
		result = append(result, i18nItem{
			Name: item.GetName(),
			URL:  i18n.Get(lg, "item.url", i18n.Vars{"id": item.GetId()}),
		})
	}

	return result
}
