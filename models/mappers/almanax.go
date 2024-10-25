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

func MapAlmanaxResourceRequest(duration int32, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		Game:     constants.GetGame().AMQPGame,
		EncyclopediaAlmanaxResourceRequest: &amqp.EncyclopediaAlmanaxResourceRequest{
			Duration: duration,
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

func MapAlmanaxEffectRequest(query string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		Game:     constants.GetGame().AMQPGame,
		EncyclopediaAlmanaxEffectRequest: &amqp.EncyclopediaAlmanaxEffectRequest{
			Query: query,
		},
	}
}

func MapAlmanaxToWebhook(almanax *amqp.Almanax, missingAlmanaxKey string,
	lg discordgo.Locale, emojiService emojis.Service) *discordgo.WebhookEdit {
	if almanax == nil {
		content := i18n.Get(lg, missingAlmanaxKey)
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

func MapAlmanaxResourceToWebhook(almanaxResources *amqp.EncyclopediaAlmanaxResourceAnswer,
	characterNumber int32, lg discordgo.Locale, emojiService emojis.Service) *discordgo.WebhookEdit {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, int(almanaxResources.Duration))
	return &discordgo.WebhookEdit{
		Embeds: MapAlmanaxResourceToEmbeds(almanaxResources, startDate, endDate,
			characterNumber, lg, emojiService),
		Components: mapAlmanaxResourceToComponents(startDate, endDate, characterNumber, lg),
	}
}

func MapAlmanaxResourceToEmbeds(almanaxResources *amqp.EncyclopediaAlmanaxResourceAnswer,
	startDate, endDate time.Time, characterNumber int32, lg discordgo.Locale,
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
		Quantity int32
	}

	i18nTributes := make([]i18nTribute, 0)
	for _, tribute := range almanaxResources.Tributes {
		i18nTributes = append(i18nTributes, i18nTribute{
			Name:     tribute.GetItemName(),
			Emoji:    emojiService.GetItemTypeStringEmoji(tribute.GetItemType()),
			Quantity: tribute.GetQuantity() * characterNumber,
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

func mapAlmanaxResourceToComponents(startDate, endDate time.Time, characterNumber int32,
	lg discordgo.Locale) *[]discordgo.MessageComponent {
	customID := contract.CraftAlmanaxResourceCustomID(startDate, endDate)
	characterNumbers := make([]discordgo.SelectMenuOption, 0)
	for _, number := range constants.GetCharacterNumbers() {
		characterNumbers = append(characterNumbers, discordgo.SelectMenuOption{
			Label: i18n.Get(lg, "almanax.resource.character", i18n.Vars{
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
					CustomID:    customID,
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: i18n.Get(lg, "almanax.resource.placeholder"),
					Options:     characterNumbers,
				},
			},
		},
	}
}
