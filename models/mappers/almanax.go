package mappers

import (
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapAlmanaxRequest(date *time.Time, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		Game:     constants.GetGame().AMQPGame,
		EncyclopediaAlmanaxRequest: &amqp.EncyclopediaAlmanaxRequest{
			Date: timestamppb.New(*date),
		},
	}
}

func MapAlmanaxResourceRequest(duration int64, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		Game:     constants.GetGame().AMQPGame,
		EncyclopediaAlmanaxResourceRequest: &amqp.EncyclopediaAlmanaxResourceRequest{
			Duration: int32(duration),
		},
	}
}

func MapAlmanaxEffectListRequest(query string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		Game:     constants.GetGame().AMQPGame,
		EncyclopediaListRequest: &amqp.EncyclopediaListRequest{
			Query: query,
			Type:  amqp.EncyclopediaListRequest_ALMANAX_EFFECT,
		},
	}
}

func MapAlmanaxEffectRequest(query *string, date *time.Time, lg discordgo.Locale) *amqp.RabbitMQMessage {
	var effectRequest amqp.EncyclopediaAlmanaxEffectRequest
	if query != nil {
		effectRequest = amqp.EncyclopediaAlmanaxEffectRequest{
			Query: *query,
			Type:  amqp.EncyclopediaAlmanaxEffectRequest_QUERY,
		}
	} else if date != nil {
		effectRequest = amqp.EncyclopediaAlmanaxEffectRequest{
			Date: timestamppb.New(*date),
			Type: amqp.EncyclopediaAlmanaxEffectRequest_DATE,
		}
	} else {
		return nil
	}

	return &amqp.RabbitMQMessage{
		Type:                             amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_REQUEST,
		Language:                         constants.MapDiscordLocale(lg),
		Game:                             constants.GetGame().AMQPGame,
		EncyclopediaAlmanaxEffectRequest: &effectRequest,
	}
}

func MapAlmanaxToWebhook(almanax *amqp.Almanax, lg discordgo.Locale,
	emojiService emojis.Service) *discordgo.WebhookEdit {
	if almanax == nil {
		content := i18n.Get(lg, "almanax.day.missing")
		return &discordgo.WebhookEdit{
			Content: &content,
		}
	}

	return &discordgo.WebhookEdit{
		Embeds:     mapAlmanaxToEmbeds(almanax, lg, emojiService),
		Components: mapAlmanaxToComponents(almanax, lg, emojiService),
	}
}

func mapAlmanaxToEmbeds(almanax *amqp.Almanax, lg discordgo.Locale,
	emojiService emojis.Service) *[]*discordgo.MessageEmbed {
	season := constants.GetSeason(almanax.Date.AsTime())

	return &[]*discordgo.MessageEmbed{
		{
			Title: i18n.Get(lg, "almanax.day.title", i18n.Vars{"date": almanax.GetDate().Seconds}),
			URL: i18n.Get(lg, "almanax.day.url", i18n.Vars{
				"date": almanax.Date.AsTime().Format(constants.KrosmozAlmanaxDateFormat),
			}),
			Color:     season.Color,
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: season.AlmanaxIcon},
			Image:     &discordgo.MessageEmbedImage{URL: almanax.Tribute.Item.Icon},
			Author: &discordgo.MessageEmbedAuthor{
				Name:    almanax.Source.Name,
				URL:     almanax.Source.Url,
				IconURL: almanax.Source.Icon,
			},
			Footer: discord.BuildDefaultFooter(lg),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  i18n.Get(lg, "almanax.day.bonus.title"),
					Value: almanax.Bonus,
				},
				{
					Name: i18n.Get(lg, "almanax.day.tribute.title"),
					Value: i18n.Get(lg, "almanax.day.tribute.description", i18n.Vars{
						"item":     almanax.Tribute.Item.Name,
						"emoji":    emojiService.GetItemTypeStringEmoji(almanax.GetTribute().Item.GetType()),
						"quantity": almanax.Tribute.Quantity,
					}),
				},
				{
					Name: i18n.Get(lg, "almanax.day.reward.title"),
					Value: i18n.Get(lg, "almanax.day.reward.description", i18n.Vars{
						"reward":   translators.FormatNumber(almanax.Reward, lg),
						"kamaIcon": emojiService.GetMiscStringEmoji(constants.EmojiIDKama),
					}),
				},
			},
		},
	}
}

func mapAlmanaxToComponents(almanax *amqp.Almanax, lg discordgo.Locale,
	emojiService emojis.Service) *[]discordgo.MessageComponent {
	previousDate := almanax.Date.AsTime().AddDate(0, 0, -1)
	nextDate := almanax.Date.AsTime().AddDate(0, 0, 1)
	return &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: contract.CraftAlmanaxDayCustomID(previousDate),
					Label:    i18n.Get(lg, "almanax.day.previous"),
					Style:    discordgo.PrimaryButton,
					Emoji:    emojiService.GetMiscEmoji(constants.EmojiIDPrevious),
				},
				discordgo.Button{
					CustomID: contract.CraftAlmanaxDayCustomID(nextDate),
					Label:    i18n.Get(lg, "almanax.day.next"),
					Style:    discordgo.PrimaryButton,
					Emoji:    emojiService.GetMiscEmoji(constants.EmojiIDNext),
				},
			},
		},
	}
}

func MapAlmanaxEffectsToWebhook(answer *amqp.EncyclopediaAlmanaxEffectAnswer, page int,
	lg discordgo.Locale, emojiService emojis.Service) *discordgo.WebhookEdit {
	if len(answer.GetAlmanaxes()) == 0 {
		content := i18n.Get(lg, "almanax.effect.missing")
		return &discordgo.WebhookEdit{
			Content: &content,
		}
	}

	pages := len(answer.GetAlmanaxes()) / constants.MaxAlmanaxEffectPerEmbed
	if len(answer.GetAlmanaxes())%constants.MaxAlmanaxEffectPerEmbed != 0 {
		pages++
	}

	// Trick to store effect ID in customID based on day
	dayWithWantedEffect := answer.GetAlmanaxes()[0].Date.AsTime()
	crafter := func(page int) string {
		return contract.CraftAlmanaxEffectCustomID(dayWithWantedEffect, page)
	}

	return &discordgo.WebhookEdit{
		Embeds:     mapAlmanaxEffectsToEmbeds(answer, page, pages, lg),
		Components: discord.GetPaginationButtons(page, pages, crafter, lg, emojiService),
	}
}

func mapAlmanaxEffectsToEmbeds(answer *amqp.EncyclopediaAlmanaxEffectAnswer,
	page, pages int, lg discordgo.Locale) *[]*discordgo.MessageEmbed {
	offset := constants.MaxAlmanaxEffectPerEmbed * page
	almanaxFields := make([]*discordgo.MessageEmbedField, 0)
	for i := offset; i < constants.MaxAlmanaxEffectPerEmbed+offset && i < len(answer.GetAlmanaxes()); i++ {
		almanaxFields = append(almanaxFields, &discordgo.MessageEmbedField{
			Name: i18n.Get(lg, "almanax.effect.day", i18n.Vars{
				"date": answer.GetAlmanaxes()[i].GetDate().Seconds,
			}),
			Value:  answer.GetAlmanaxes()[i].GetBonus(),
			Inline: false,
		})
	}

	return &[]*discordgo.MessageEmbed{
		{
			Title: i18n.Get(lg, "almanax.effect.title", i18n.Vars{
				"query": answer.GetQuery(),
			}),
			Description: i18n.Get(lg, "almanax.effect.description", i18n.Vars{
				"len":   len(answer.GetAlmanaxes()),
				"page":  page + 1,
				"pages": pages,
			}),
			Color: constants.Color,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: constants.GetUnknownSeason().AlmanaxIcon,
			},
			Author: &discordgo.MessageEmbedAuthor{
				Name:    answer.GetAlmanaxes()[offset].Source.Name,
				URL:     answer.GetAlmanaxes()[offset].Source.Url,
				IconURL: answer.GetAlmanaxes()[offset].Source.Icon,
			},
			Fields: almanaxFields,
			Footer: discord.BuildDefaultFooter(lg),
		},
	}
}

func MapAlmanaxResourceToWebhook(almanaxResources *amqp.EncyclopediaAlmanaxResourceAnswer,
	characterNumber int64, lg discordgo.Locale, emojiService emojis.Service) *discordgo.WebhookEdit {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, int(almanaxResources.Duration))
	return &discordgo.WebhookEdit{
		Embeds: mapAlmanaxResourceToEmbeds(almanaxResources, startDate, endDate,
			characterNumber, lg, emojiService),
		Components: mapAlmanaxResourceToComponents(int64(almanaxResources.Duration), characterNumber, lg),
	}
}

func mapAlmanaxResourceToEmbeds(almanaxResources *amqp.EncyclopediaAlmanaxResourceAnswer,
	startDate, endDate time.Time, characterNumber int64, lg discordgo.Locale,
	emojiService emojis.Service) *[]*discordgo.MessageEmbed {
	collator := constants.MapCollator(lg)
	sort.SliceStable(almanaxResources.Tributes, func(i, j int) bool {
		if almanaxResources.Tributes[i].ItemType == almanaxResources.Tributes[j].ItemType {
			return collator.CompareString(almanaxResources.Tributes[i].ItemName,
				almanaxResources.Tributes[j].ItemName) == -1
		}

		return almanaxResources.Tributes[i].ItemType < almanaxResources.Tributes[j].ItemType
	})

	type i18nTribute struct {
		Name     string
		Emoji    string
		Quantity int64
	}

	i18nTributes := make([]i18nTribute, 0)
	for _, tribute := range almanaxResources.Tributes {
		i18nTributes = append(i18nTributes, i18nTribute{
			Name:     tribute.GetItemName(),
			Emoji:    emojiService.GetItemTypeStringEmoji(tribute.GetItemType()),
			Quantity: int64(tribute.GetQuantity()) * characterNumber,
		})
	}

	return &[]*discordgo.MessageEmbed{
		{
			Title: i18n.Get(lg, "almanax.resource.title", i18n.Vars{
				"startDate": startDate.Unix(),
				"endDate":   endDate.Unix(),
			}),
			Description: i18n.Get(lg, "almanax.resource.description", i18n.Vars{
				"number":   characterNumber,
				"tributes": i18nTributes,
			}),
			Color:     constants.Color,
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: constants.GetUnknownSeason().AlmanaxIcon},
			Author: &discordgo.MessageEmbedAuthor{
				Name:    almanaxResources.Source.Name,
				URL:     almanaxResources.Source.Url,
				IconURL: almanaxResources.Source.Icon,
			},
			Footer: discord.BuildDefaultFooter(lg),
		},
	}
}

func mapAlmanaxResourceToComponents(duration, characterNumber int64,
	lg discordgo.Locale) *[]discordgo.MessageComponent {
	durationCustomID := contract.CraftAlmanaxResourceDurationCustomID(characterNumber)
	durationValues := make([]discordgo.SelectMenuOption, 0)
	for _, number := range constants.GetAlmanaxDayDuration() {
		durationValues = append(durationValues, discordgo.SelectMenuOption{
			Label: i18n.Get(lg, "almanax.resource.duration.label", i18n.Vars{
				"number": number,
			}),
			Value:   fmt.Sprintf("%v", number),
			Default: duration == number,
		})
	}

	characterCustomID := contract.CraftAlmanaxResourceCharacterCustomID(duration)
	characterNumbers := make([]discordgo.SelectMenuOption, 0)
	for _, number := range constants.GetCharacterNumbers() {
		characterNumbers = append(characterNumbers, discordgo.SelectMenuOption{
			Label: i18n.Get(lg, "almanax.resource.character.label", i18n.Vars{
				"number": number,
			}),
			Value:   fmt.Sprintf("%v", number),
			Default: characterNumber == number,
		})
	}

	return &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    durationCustomID,
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: i18n.Get(lg, "almanax.resource.duration.placeholder"),
					Options:     durationValues,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    characterCustomID,
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: i18n.Get(lg, "almanax.resource.character.placeholder"),
					Options:     characterNumbers,
				},
			},
		},
	}
}
