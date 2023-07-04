package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func MapItemListRequest(query string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_LIST_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaItemListRequest: &amqp.EncyclopediaItemListRequest{
			Query: query,
		},
	}
}

func MapItemRequest(query string, isID bool, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaItemRequest: &amqp.EncyclopediaItemRequest{
			Query: query,
			IsID:  isID,
		},
	}
}

func MapItemToWebhookEdit(item *amqp.EncyclopediaItemAnswer, isRecipe bool,
	characService characteristics.Service, emojiService emojis.Service,
	locale amqp.Language) *discordgo.WebhookEdit {
	lg := constants.MapAMQPLocale(locale)

	return &discordgo.WebhookEdit{
		Embeds:     mapItemToEmbeds(item, isRecipe, characService, lg),
		Components: mapItemToComponents(item, isRecipe, emojiService, lg),
	}
}

func mapItemToEmbeds(item *amqp.EncyclopediaItemAnswer, isRecipe bool,
	service characteristics.Service, lg discordgo.Locale) *[]*discordgo.MessageEmbed {
	fields := make([]*discordgo.MessageEmbedField, 0)

	if !isRecipe && len(item.GetEffects()) > 0 {
		i18nEffects := mapEffects(item.GetEffects(), service)
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
	} else if isRecipe && item.GetRecipe() != nil {
		recipeFields := discord.SliceFields(item.GetRecipe().GetIngredients(), constants.MaxIngredientsPerField,
			func(i int, items []*amqp.EncyclopediaItemAnswer_Ingredient) *discordgo.MessageEmbedField {
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
			Title: item.GetName(),
			Description: i18n.Get(lg, "item.description", i18n.Vars{
				"level":       item.GetLevel(),
				"type":        item.GetLabelType(),
				"description": item.GetDescription(),
			}),
			Color: constants.Color,
			URL:   i18n.Get(lg, "item.url", i18n.Vars{"id": item.GetId()}),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: item.GetIcon(),
			},
			Fields: fields,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    item.GetSource().GetName(),
				URL:     item.GetSource().GetUrl(),
				IconURL: item.GetSource().GetIcon(),
			},
		},
	}
}

func mapItemToComponents(item *amqp.EncyclopediaItemAnswer, isRecipe bool,
	service emojis.Service, lg discordgo.Locale) *[]discordgo.MessageComponent {
	components := make([]discordgo.MessageComponent, 0)

	if item.GetSet() != nil {
		components = append(components, discordgo.Button{
			CustomID: contract.CraftSetCustomID(item.GetSet().GetId()),
			Label:    item.GetSet().GetName(),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDSet),
		})
	}

	if isRecipe && len(item.GetEffects()) > 0 {
		components = append(components, discordgo.Button{
			CustomID: contract.CraftItemEffectsCustomID(item.GetId()),
			Label:    i18n.Get(lg, "item.effects.button"),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDEffect),
		})
	} else if item.GetRecipe() != nil {
		components = append(components, discordgo.Button{
			CustomID: contract.CraftItemRecipeCustomID(item.GetId()),
			Label:    i18n.Get(lg, "item.recipe.button"),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDRecipe),
		})
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

func mapItemIngredients(ingredients []*amqp.EncyclopediaItemAnswer_Ingredient,
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
