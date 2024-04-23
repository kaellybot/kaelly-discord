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

func MapItemListRequest(query string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaListRequest: &amqp.EncyclopediaListRequest{
			Query: query,
			Type:  amqp.EncyclopediaListRequest_ITEM,
		},
	}
}

func MapItemRequest(query string, isID bool, itemType amqp.ItemType,
	lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaItemRequest: &amqp.EncyclopediaItemRequest{
			Query: query,
			IsID:  isID,
			Type:  itemType,
		},
	}
}

//nolint:nolintlint,exhaustive
func MapItemToWebhookEdit(answer *amqp.EncyclopediaItemAnswer, isRecipe bool,
	characService characteristics.Service, emojiService emojis.Service,
	locale amqp.Language) (*discordgo.WebhookEdit, error) {
	// TODO handle all these types
	switch answer.GetType() {
	case amqp.ItemType_CONSUMABLE:
		return nil, ErrItemTypeNotHandled
	case amqp.ItemType_COSMETIC:
		return nil, ErrItemTypeNotHandled
	case amqp.ItemType_EQUIPMENT:
		return mapEquipmentToWebhookEdit(answer, isRecipe, characService,
			emojiService, locale), nil
	case amqp.ItemType_MOUNT:
		return nil, ErrItemTypeNotHandled
	case amqp.ItemType_QUEST_ITEM:
		return nil, ErrItemTypeNotHandled
	case amqp.ItemType_RESOURCE:
		return nil, ErrItemTypeNotHandled
	default:
		return nil, ErrItemTypeNotHandled
	}
}

func mapEquipmentToWebhookEdit(answer *amqp.EncyclopediaItemAnswer, isRecipe bool,
	characService characteristics.Service, emojiService emojis.Service,
	locale amqp.Language) *discordgo.WebhookEdit {
	lg := constants.MapAMQPLocale(locale)
	return &discordgo.WebhookEdit{
		Embeds:     mapEquipmentToEmbeds(answer, isRecipe, characService, lg),
		Components: mapEquipmentToComponents(answer, isRecipe, emojiService, lg),
	}
}

func mapEquipmentToEmbeds(answer *amqp.EncyclopediaItemAnswer, isRecipe bool,
	service characteristics.Service, lg discordgo.Locale) *[]*discordgo.MessageEmbed {
	equipment := answer.GetEquipment()
	fields := make([]*discordgo.MessageEmbedField, 0)

	if !isRecipe && len(equipment.GetEffects()) > 0 {
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
	} else if isRecipe && equipment.GetRecipe() != nil {
		recipeFields := discord.SliceFields(equipment.GetRecipe().GetIngredients(), constants.MaxIngredientsPerField,
			func(i int, items []*amqp.EncyclopediaItemAnswer_Recipe_Ingredient) *discordgo.MessageEmbedField {
				name := constants.InvisibleCharacter
				if i == 0 {
					name = i18n.Get(lg, "item.recipe.title")
				}

				return &discordgo.MessageEmbedField{
					Name: name,
					Value: i18n.Get(lg, "item.recipe.description", i18n.Vars{
						"ingredients": mapItemIngredients(items, lg),
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
				"level":       equipment.GetLevel(),
				"type":        equipment.GetLabelType(),
				"description": equipment.GetDescription(),
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
			Emoji:    service.GetMiscEmoji(constants.EmojiIDSet),
		})
	}

	if isRecipe && len(equipment.GetEffects()) > 0 {
		components = append(components, discordgo.Button{
			CustomID: contract.CraftItemEffectsCustomID(equipment.GetId(), amqp.ItemType_EQUIPMENT.String()),
			Label:    i18n.Get(lg, "item.effects.button"),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDEffect),
		})
	} else if equipment.GetRecipe() != nil {
		components = append(components, discordgo.Button{
			CustomID: contract.CraftItemRecipeCustomID(equipment.GetId(), amqp.ItemType_EQUIPMENT.String()),
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
	Quantity int64
}

func mapItemIngredients(ingredients []*amqp.EncyclopediaItemAnswer_Recipe_Ingredient,
	lg discordgo.Locale) []i18nIngredient {
	result := make([]i18nIngredient, 0)
	for _, ingredient := range ingredients {
		result = append(result, i18nIngredient{
			Name:     ingredient.GetName(),
			URL:      i18n.Get(lg, "item.url", i18n.Vars{"id": ingredient.GetId()}), // TODO not so simple
			Quantity: ingredient.GetQuantity(),
		})
	}

	return result
}
