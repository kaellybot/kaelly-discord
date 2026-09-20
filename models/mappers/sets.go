package mappers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	di18n "github.com/kaysoro/discordgo-i18n"
)

func MapSetListRequest(query, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, lg)
	request.EncyclopediaListRequest = &amqp.EncyclopediaListRequest{
		Query: query,
		Type:  amqp.EncyclopediaListRequest_SET,
	}
	return request
}

func MapSetToDefaultWebhookEdit(answer *amqp.EncyclopediaItemAnswer,
	characService characteristics.Service, emojiService emojis.Service,
	locale amqp.Language) *discordgo.WebhookEdit {
	return MapSetToWebhookEdit(answer, len(answer.GetSet().GetEquipments()),
		characService, emojiService, locale)
}

func MapSetToWebhookEdit(answer *amqp.EncyclopediaItemAnswer, itemNumber int,
	characService characteristics.Service, emojiService emojis.Service,
	locale amqp.Language) *discordgo.WebhookEdit {
	lg := i18n.MapAMQPLocale(locale)
	bonus := findSetBonus(answer.GetSet().GetBonuses(), int64(itemNumber))

	return &discordgo.WebhookEdit{
		Embeds:     mapSetToEmbeds(answer, bonus, characService, emojiService, lg),
		Components: mapSetToComponents(answer, bonus, emojiService, lg),
	}
}

// findSetBonus returns the bonus matching itemNumber, else the closest bonus
// below it, else the lowest one; nil when the set has no bonus at all.
func findSetBonus(bonuses []*amqp.EncyclopediaItemAnswer_Set_Bonus,
	itemNumber int64) *amqp.EncyclopediaItemAnswer_Set_Bonus {
	var closest, lowest *amqp.EncyclopediaItemAnswer_Set_Bonus
	for _, bonus := range bonuses {
		currentNumber := bonus.GetItemNumber()
		if currentNumber == itemNumber {
			return bonus
		}

		if currentNumber < itemNumber && (closest == nil || currentNumber > closest.GetItemNumber()) {
			closest = bonus
		}

		if lowest == nil || currentNumber < lowest.GetItemNumber() {
			lowest = bonus
		}
	}

	if closest != nil {
		return closest
	}

	return lowest
}

func mapSetToEmbeds(answer *amqp.EncyclopediaItemAnswer,
	bonus *amqp.EncyclopediaItemAnswer_Set_Bonus, service characteristics.Service,
	emojiService emojis.Service, lg discordgo.Locale) *[]*discordgo.MessageEmbed {
	set := answer.GetSet()
	fields := discord.SliceFields(set.GetEquipments(), constants.MaxEquipmentPerField,
		func(i int, items []*amqp.EncyclopediaItemAnswer_Set_Equipment) *discordgo.MessageEmbedField {
			name := constants.InvisibleCharacter
			if i == 0 {
				name = di18n.Get(lg, "set.items.title")
			}

			return &discordgo.MessageEmbedField{
				Name: name,
				Value: di18n.Get(lg, "set.items.description", di18n.Vars{
					"items": mapSetItems(items),
				}),
				Inline: true,
			}
		})

	if bonus != nil {
		i18nEffects := mapEffects(bonus.GetEffects(), service, emojiService)
		bonusFields := discord.SliceFields(i18nEffects, constants.MaxCharacterPerField,
			func(i int, items []i18nCharacteristic) *discordgo.MessageEmbedField {
				name := constants.InvisibleCharacter
				if i == 0 {
					name = di18n.Get(lg, "set.effects.title", di18n.Vars{
						"itemNumber": bonus.GetItemNumber(),
					})
				}

				return &discordgo.MessageEmbedField{
					Name: name,
					Value: di18n.Get(lg, "set.effects.description", di18n.Vars{
						"effects": items,
					}),
					Inline: true,
				}
			})

		if len(bonusFields) > 0 {
			fields = append(fields, discord.GhostInlineField())
			fields = append(fields, bonusFields...)
		}
	}

	return &[]*discordgo.MessageEmbed{
		{
			Title:       set.GetName(),
			Description: di18n.Get(lg, "set.description", di18n.Vars{"level": set.GetLevel()}),
			Color:       constants.Color,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: set.GetIcon()},
			Fields:      fields,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    answer.GetSource().GetName(),
				URL:     answer.GetSource().GetUrl(),
				IconURL: answer.GetSource().GetIcon(),
			},
			Footer: discord.BuildDefaultFooter(lg),
		},
	}
}

func mapSetToComponents(answer *amqp.EncyclopediaItemAnswer,
	bonus *amqp.EncyclopediaItemAnswer_Set_Bonus, service emojis.Service,
	lg discordgo.Locale) *[]discordgo.MessageComponent {
	set := answer.GetSet()
	components := make([]discordgo.MessageComponent, 0)

	var maxItemNumber int
	for _, currentBonus := range set.GetBonuses() {
		itemNumber := int(currentBonus.GetItemNumber())
		if itemNumber > maxItemNumber {
			maxItemNumber = itemNumber
		}
	}
	bonuses := make([]discordgo.SelectMenuOption, 0)
	for _, currentBonus := range set.GetBonuses() {
		emoji := service.GetSetBonusEmoji(int(currentBonus.GetItemNumber()))
		bonuses = append(bonuses, discordgo.SelectMenuOption{
			Label: di18n.Get(lg, "set.effects.option", di18n.Vars{
				"itemNumber": currentBonus.GetItemNumber(),
				"itemCount":  maxItemNumber,
			}),
			Value:   fmt.Sprintf("%v", currentBonus.GetItemNumber()),
			Default: bonus != nil && currentBonus.GetItemNumber() == bonus.GetItemNumber(),
			Emoji:   emoji,
		})
	}

	if len(bonuses) > 0 {
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    contract.CraftSetBonusCustomID(set.GetId()),
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: di18n.Get(lg, "set.effects.placeholder"),
					Options:     bonuses,
				},
			},
		})
	}

	items := make([]discordgo.SelectMenuOption, 0)
	for _, item := range set.GetEquipments() {
		items = append(items, discordgo.SelectMenuOption{
			Label: item.GetName(),
			Value: item.GetId(),
			Emoji: service.GetEquipmentEmoji(item.GetType()),
		})
	}

	var itemType amqp.ItemType
	if set.GetIsCosmetic() {
		itemType = amqp.ItemType_COSMETIC_TYPE
	} else {
		itemType = amqp.ItemType_EQUIPMENT_TYPE
	}

	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    contract.CraftItemCustomID(itemType.String()),
				MenuType:    discordgo.StringSelectMenu,
				Placeholder: di18n.Get(lg, "set.items.placeholder"),
				Options:     items,
			},
		},
	})

	return &components
}

type i18nSetItem struct {
	Name  string
	Level int64
}

func mapSetItems(items []*amqp.EncyclopediaItemAnswer_Set_Equipment) []i18nSetItem {
	result := make([]i18nSetItem, 0)
	for _, item := range items {
		result = append(result, i18nSetItem{
			Name:  item.GetName(),
			Level: item.GetLevel(),
		})
	}

	return result
}
