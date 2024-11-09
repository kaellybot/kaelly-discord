package mappers

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
)

var (
	ErrItemTypeNotHandled = errors.New("item type not handled")
)

func MapItemListRequest(query, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, lg)
	request.EncyclopediaListRequest = &amqp.EncyclopediaListRequest{
		Query: query,
		Type:  amqp.EncyclopediaListRequest_ITEM,
	}
	return request
}

func MapItemRequest(query string, isID bool, itemType amqp.ItemType,
	authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_REQUEST, lg)
	request.EncyclopediaItemRequest = &amqp.EncyclopediaItemRequest{
		Query: query,
		IsID:  isID,
		Type:  itemType,
	}
	return request
}

func MapItemToWebhookEdit(answer *amqp.EncyclopediaItemAnswer, isRecipe bool,
	characService characteristics.Service, emojiService emojis.Service,
	locale amqp.Language) *discordgo.WebhookEdit {
	return mapEquipmentToWebhookEdit(answer, isRecipe, characService,
		emojiService, locale)
}

func mapEquipmentToWebhookEdit(answer *amqp.EncyclopediaItemAnswer, isRecipe bool,
	characService characteristics.Service, emojiService emojis.Service,
	locale amqp.Language) *discordgo.WebhookEdit {
	lg := constants.MapAMQPLocale(locale)
	return &discordgo.WebhookEdit{
		Embeds:     mapEquipmentToEmbeds(answer, isRecipe, characService, emojiService, lg),
		Components: mapEquipmentToComponents(answer, isRecipe, emojiService, lg),
	}
}

func mapEquipmentToEmbeds(answer *amqp.EncyclopediaItemAnswer, isRecipe bool,
	service characteristics.Service, emojiService emojis.Service, lg discordgo.Locale,
) *[]*discordgo.MessageEmbed {
	equipment := answer.GetEquipment()
	fields := make([]*discordgo.MessageEmbedField, 0)

	if !isRecipe && equipment.GetCharacteristics() != nil {
		characteristics := equipment.GetCharacteristics()
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: i18n.Get(lg, "item.characteristics.title"),
			Value: i18n.Get(lg, "item.characteristics.description", i18n.Vars{
				"cost":           characteristics.GetCost(),
				"costEmoji":      emojiService.GetMiscStringEmoji(constants.EmojiIDCost),
				"minRange":       characteristics.GetMinRange(),
				"maxRange":       characteristics.GetMaxRange(),
				"rangeEmoji":     emojiService.GetMiscStringEmoji(constants.EmojiIDRange),
				"maxCastPerTurn": characteristics.GetMaxCastPerTurn(),
				"criticalRate":   characteristics.GetCriticalRate(),
				"criticalBonus":  characteristics.GetCriticalBonus(),
				"criticalEmoji":  emojiService.GetMiscStringEmoji(constants.EmojiIDCritical),
				// TODO area + LDV
			}),
			Inline: false,
		})
	}

	if !isRecipe && (len(equipment.GetWeaponEffects()) > 0 || len(equipment.GetEffects()) > 0) {
		i18nWeaponEffects := mapEffects(equipment.GetWeaponEffects(), service)
		weaponEffectFields := discord.SliceFields(i18nWeaponEffects, constants.MaxCharacterPerField,
			func(i int, items []i18nCharacteristic) *discordgo.MessageEmbedField {
				name := constants.InvisibleCharacter
				if i == 0 {
					name = i18n.Get(lg, "item.weaponEffects.title")
				}

				return &discordgo.MessageEmbedField{
					Name: name,
					Value: i18n.Get(lg, "item.weaponEffects.description", i18n.Vars{
						"effects": items,
					}),
					Inline: false,
				}
			})
		fields = append(fields, weaponEffectFields...)

		i18nEffects := mapEffects(equipment.GetEffects(), service)
		effectFields := discord.SliceFields(i18nEffects, constants.MaxCharacterPerField,
			func(i int, items []i18nCharacteristic) *discordgo.MessageEmbedField {
				name := constants.InvisibleCharacter
				if i == 0 {
					name = i18n.Get(lg, "item.effects.title")
				}

				return &discordgo.MessageEmbedField{
					Name: name,
					Value: i18n.Get(lg, "item.effects.description", i18n.Vars{
						"effects": items,
					}),
					Inline: true,
				}
			})
		fields = append(fields, effectFields...)
	}

	if !isRecipe && len(equipment.Conditions) > 0 {
		// TODO add field conditions
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "",
			Value:  "",
			Inline: false,
		})
	}

	if isRecipe && equipment.GetRecipe() != nil {
		recipeFields := discord.SliceFields(equipment.GetRecipe().GetIngredients(), constants.MaxIngredientsPerField,
			func(i int, items []*amqp.EncyclopediaItemAnswer_Recipe_Ingredient) *discordgo.MessageEmbedField {
				name := constants.InvisibleCharacter
				if i == 0 {
					name = i18n.Get(lg, "item.recipe.title")
				}

				return &discordgo.MessageEmbedField{
					Name: name,
					Value: i18n.Get(lg, "item.recipe.description", i18n.Vars{
						"ingredients": mapItemIngredients(items, emojiService, lg),
					}),
					Inline: true,
				}
			})

		fields = append(fields, recipeFields...)
	}

	return &[]*discordgo.MessageEmbed{
		{
			Title: equipment.GetName(),
			Description: i18n.Get(lg, "item.description", i18n.Vars{
				"level": equipment.GetLevel(),
				"emoji": emojiService.GetEquipmentStringEmoji(equipment.GetType().GetEquipmentType()),
				"type":  equipment.GetType().GetEquipmentLabel(),
			}),
			Color: constants.Color,
			URL:   i18n.Get(lg, "item.url", i18n.Vars{"id": equipment.GetId()}),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: equipment.GetIcon(),
			},
			Fields: fields,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    answer.GetSource().GetName(),
				URL:     answer.GetSource().GetUrl(),
				IconURL: answer.GetSource().GetIcon(),
			},
			Footer: discord.BuildDefaultFooter(lg),
		},
	}
}

func mapEquipmentToComponents(answer *amqp.EncyclopediaItemAnswer, isRecipe bool,
	service emojis.Service, lg discordgo.Locale) *[]discordgo.MessageComponent {
	equipment := answer.GetEquipment()
	components := make([]discordgo.MessageComponent, 0)

	if equipment.GetSet() != nil {
		components = append(components, discordgo.Button{
			CustomID: contract.CraftSetCustomID(equipment.GetSet().GetId()),
			Label:    equipment.GetSet().GetName(),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetItemTypeEmoji(amqp.ItemType_SET_TYPE),
		})
	}

	if isRecipe && (len(equipment.GetWeaponEffects()) > 0 || len(equipment.GetEffects()) > 0) {
		components = append(components, discordgo.Button{
			CustomID: contract.CraftItemEffectsCustomID(equipment.GetId(), amqp.ItemType_EQUIPMENT_TYPE.String()),
			Label:    i18n.Get(lg, "item.effects.button"),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDEffect),
		})
	} else if equipment.GetRecipe() != nil {
		components = append(components, discordgo.Button{
			CustomID: contract.CraftItemRecipeCustomID(equipment.GetId(), amqp.ItemType_EQUIPMENT_TYPE.String()),
			Label:    i18n.Get(lg, "item.recipe.button"),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDRecipe),
		})
	}

	if len(components) == 0 {
		return nil
	}

	return &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: components,
		},
	}
}

type i18nIngredient struct {
	Name     string
	URL      string
	Emoji    string
	Quantity int64
}

func mapItemIngredients(ingredients []*amqp.EncyclopediaItemAnswer_Recipe_Ingredient,
	emojiService emojis.Service, lg discordgo.Locale) []i18nIngredient {
	result := make([]i18nIngredient, 0)
	for _, ingredient := range ingredients {
		result = append(result, i18nIngredient{
			Name:     ingredient.GetName(),
			URL:      i18n.Get(lg, "item.url", i18n.Vars{"id": ingredient.GetId()}),
			Emoji:    emojiService.GetItemTypeStringEmoji(ingredient.GetType()),
			Quantity: ingredient.GetQuantity(),
		})
	}

	return result
}
