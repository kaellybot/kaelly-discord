package mappers

import (
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	di18n "github.com/kaysoro/discordgo-i18n"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapAlmanaxRequest(date *time.Time, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_REQUEST, lg)
	request.EncyclopediaAlmanaxRequest = &amqp.EncyclopediaAlmanaxRequest{
		Date: timestamppb.New(*date),
	}
	return request
}

func MapAlmanaxResourceRequest(duration int64, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_REQUEST, lg)
	request.EncyclopediaAlmanaxResourceRequest = &amqp.EncyclopediaAlmanaxResourceRequest{
		Duration: duration,
	}
	return request
}

func MapAlmanaxEffectListRequest(query, authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, lg)
	request.EncyclopediaListRequest = &amqp.EncyclopediaListRequest{
		Query: query,
		Type:  amqp.EncyclopediaListRequest_ALMANAX_EFFECT,
	}
	return request
}

func MapAlmanaxEffectRequest(query *string, date *time.Time, page int, authorID string,
	lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_REQUEST, lg)
	effectRequest := amqp.EncyclopediaAlmanaxEffectRequest{
		Offset: int64(page) * constants.MaxAlmanaxEffectPerEmbed,
		Size:   constants.MaxAlmanaxEffectPerEmbed,
	}

	switch {
	case query != nil:
		effectRequest.Query = *query
		effectRequest.Type = amqp.EncyclopediaAlmanaxEffectRequest_QUERY
	case date != nil:
		effectRequest.Date = timestamppb.New(*date)
		effectRequest.Type = amqp.EncyclopediaAlmanaxEffectRequest_DATE
	default:
		return nil
	}

	request.EncyclopediaAlmanaxEffectRequest = &effectRequest
	return request
}

func MapAlmanaxToWebhook(answer *amqp.EncyclopediaAlmanaxAnswer, lg discordgo.Locale,
	emojiService emojis.Service) *discordgo.WebhookEdit {
	if answer.GetAlmanax() == nil {
		content := di18n.Get(lg, "almanax.day.missing")
		return &discordgo.WebhookEdit{
			Content: &content,
		}
	}

	return &discordgo.WebhookEdit{
		Embeds:     mapAlmanaxToEmbeds(answer, lg, emojiService),
		Components: mapAlmanaxToComponents(answer.GetAlmanax(), lg, emojiService),
	}
}

func mapAlmanaxToEmbeds(answer *amqp.EncyclopediaAlmanaxAnswer, lg discordgo.Locale,
	emojiService emojis.Service) *[]*discordgo.MessageEmbed {
	dateTranslator := i18n.MapDateTranslator(lg)
	almanax := answer.GetAlmanax()
	season := constants.GetSeason(almanax.GetDate().AsTime())
	fullDate := dateTranslator.FmtDateFull(almanax.GetDate().AsTime())

	return &[]*discordgo.MessageEmbed{
		{
			Title: di18n.Get(lg, "almanax.day.title", di18n.Vars{"date": fullDate}),
			URL: di18n.Get(lg, "almanax.day.url", di18n.Vars{
				"date": almanax.GetDate().AsTime().Format(constants.KrosmozAlmanaxDateFormat),
			}),
			Color:     season.Color,
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: season.AlmanaxIcon},
			Image:     &discordgo.MessageEmbedImage{URL: almanax.GetTribute().GetItem().GetIcon()},
			Author: &discordgo.MessageEmbedAuthor{
				Name:    answer.GetSource().GetName(),
				URL:     answer.GetSource().GetUrl(),
				IconURL: answer.GetSource().GetIcon(),
			},
			Footer: discord.BuildDefaultFooter(lg),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  di18n.Get(lg, "almanax.day.bonus.title"),
					Value: almanax.GetBonus(),
				},
				{
					Name: di18n.Get(lg, "almanax.day.tribute.title"),
					Value: di18n.Get(lg, "almanax.day.tribute.description", di18n.Vars{
						"item":     almanax.GetTribute().GetItem().GetName(),
						"emoji":    emojiService.GetItemTypeStringEmoji(almanax.GetTribute().GetItem().GetType()),
						"quantity": almanax.GetTribute().GetQuantity(),
					}),
				},
				{
					Name: di18n.Get(lg, "almanax.day.reward.title"),
					Value: di18n.Get(lg, "almanax.day.reward.description", di18n.Vars{
						"reward":   translators.FormatNumber(almanax.GetReward(), lg),
						"kamaIcon": emojiService.GetMiscStringEmoji(constants.EmojiIDKama),
					}),
				},
			},
		},
	}
}

func mapAlmanaxToComponents(almanax *amqp.Almanax, lg discordgo.Locale,
	emojiService emojis.Service) *[]discordgo.MessageComponent {
	previousDate := almanax.GetDate().AsTime().AddDate(0, 0, -1)
	nextDate := almanax.GetDate().AsTime().AddDate(0, 0, 1)
	return &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: contract.CraftAlmanaxDayCustomID(previousDate),
					Label:    di18n.Get(lg, "almanax.day.previous"),
					Style:    discordgo.SecondaryButton,
					Emoji:    emojiService.GetMiscEmoji(constants.EmojiIDPrevious),
				},
				discordgo.Button{
					CustomID: contract.CraftAlmanaxDayCustomID(nextDate),
					Label:    di18n.Get(lg, "almanax.day.next"),
					Style:    discordgo.SecondaryButton,
					Emoji:    emojiService.GetMiscEmoji(constants.EmojiIDNext),
				},
				discordgo.Button{
					CustomID: contract.CraftAlmanaxEffectCustomID(almanax.GetDate().AsTime(), constants.DefaultPage),
					Label:    di18n.Get(lg, "almanax.day.effect"),
					Style:    discordgo.PrimaryButton,
					Emoji:    emojiService.GetMiscEmoji(constants.EmojiIDEffect),
				},
			},
		},
	}
}

func MapAlmanaxEffectsToWebhook(answer *amqp.EncyclopediaAlmanaxEffectAnswer,
	lg discordgo.Locale, emojiService emojis.Service) *discordgo.WebhookEdit {
	// Effect name was not even found
	if answer.GetEffectName() == "" {
		return MapQueryMismatch(answer.GetQuery(), lg)
	}

	// Effect found but no attached almanax days
	if len(answer.GetAlmanaxes()) == 0 {
		content := di18n.Get(lg, "almanax.effect.missing")
		return &discordgo.WebhookEdit{
			Content: &content,
		}
	}

	return &discordgo.WebhookEdit{
		Embeds:     mapAlmanaxEffectsToEmbeds(answer, lg, emojiService),
		Components: mapAlmanaxEffectsToComponents(answer, lg, emojiService),
	}
}

func mapAlmanaxEffectsToEmbeds(answer *amqp.EncyclopediaAlmanaxEffectAnswer,
	lg discordgo.Locale, emojiService emojis.Service) *[]*discordgo.MessageEmbed {
	almanaxFields := make([]*discordgo.MessageEmbedField, 0)
	for _, almanax := range answer.GetAlmanaxes() {
		almanaxFields = append(almanaxFields, &discordgo.MessageEmbedField{
			Name: di18n.Get(lg, "almanax.effect.day", di18n.Vars{
				"emoji": emojiService.GetMiscStringEmoji(constants.EmojiIDCalendar),
				"date":  almanax.GetDate().GetSeconds(),
			}),
			Value:  almanax.GetBonus(),
			Inline: false,
		})
	}

	effectName := answer.GetEffectName()
	if len(effectName) == 0 {
		effectName = answer.GetQuery()
	}

	return &[]*discordgo.MessageEmbed{
		{
			Title: di18n.Get(lg, "almanax.effect.title", di18n.Vars{
				"query": effectName,
			}),
			Description: di18n.Get(lg, "almanax.effect.description", di18n.Vars{
				"total": answer.GetTotal(),
				"page":  answer.GetPage() + 1,
				"pages": answer.GetPages(),
			}),
			Color: constants.Color,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: constants.GetUnknownSeason().AlmanaxIcon,
			},
			Author: &discordgo.MessageEmbedAuthor{
				Name:    answer.GetSource().GetName(),
				URL:     answer.GetSource().GetUrl(),
				IconURL: answer.GetSource().GetIcon(),
			},
			Fields: almanaxFields,
			Footer: discord.BuildDefaultFooter(lg),
		},
	}
}

func mapAlmanaxEffectsToComponents(answer *amqp.EncyclopediaAlmanaxEffectAnswer,
	lg discordgo.Locale, emojiService emojis.Service) *[]discordgo.MessageComponent {
	almanaxes := answer.GetAlmanaxes()
	// Trick to store effect ID in customID based on day
	dayWithWantedEffect := almanaxes[0].GetDate().AsTime()
	crafter := func(page int) string {
		return contract.CraftAlmanaxEffectCustomID(dayWithWantedEffect, page)
	}

	almanaxChoices := make([]discordgo.SelectMenuOption, 0)
	for _, almanax := range almanaxes {
		almanaxChoices = append(almanaxChoices, discordgo.SelectMenuOption{
			Label: di18n.Get(lg, "almanax.effect.choice.value", di18n.Vars{
				"date": almanax.GetDate().AsTime().Format(constants.DiscordDateOnlyFormat),
			}),
			Value: fmt.Sprintf("%v", almanax.GetDate().GetSeconds()),
			Emoji: emojiService.GetMiscEmoji(constants.EmojiIDCalendar),
		})
	}

	components := make([]discordgo.MessageComponent, 0)
	paginations := discord.GetPaginationButtons(int(answer.GetPage()), int(answer.GetPages()),
		crafter, lg, emojiService)
	if len(paginations) > 0 {
		components = append(components, discordgo.ActionsRow{
			Components: paginations,
		})
	}

	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    contract.CraftAlmanaxDayChoiceCustomID(),
				MenuType:    discordgo.StringSelectMenu,
				Placeholder: di18n.Get(lg, "almanax.effect.choice.placeholder"),
				Options:     almanaxChoices,
			},
		},
	})

	return &components
}

func MapAlmanaxResourceToWebhook(almanaxResources *amqp.EncyclopediaAlmanaxResourceAnswer,
	characterNumber int64, lg discordgo.Locale, emojiService emojis.Service) *discordgo.WebhookEdit {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, int(almanaxResources.GetDuration()))
	return &discordgo.WebhookEdit{
		Embeds: mapAlmanaxResourceToEmbeds(almanaxResources, startDate, endDate,
			characterNumber, lg, emojiService),
		Components: mapAlmanaxResourceToComponents(almanaxResources.GetDuration(), characterNumber, lg),
	}
}

func mapAlmanaxResourceToEmbeds(almanaxResources *amqp.EncyclopediaAlmanaxResourceAnswer,
	startDate, endDate time.Time, characterNumber int64, lg discordgo.Locale,
	emojiService emojis.Service) *[]*discordgo.MessageEmbed {
	collator := i18n.MapCollator(lg)
	tributes := slices.Clone(almanaxResources.GetTributes())
	sort.SliceStable(tributes, func(i, j int) bool {
		if tributes[i].GetItemType() == tributes[j].GetItemType() {
			return collator.CompareString(tributes[i].GetItemName(),
				tributes[j].GetItemName()) == -1
		}

		return tributes[i].GetItemType() < tributes[j].GetItemType()
	})

	type i18nTribute struct {
		Name     string
		Emoji    string
		Quantity int64
	}

	i18nTributes := make([]i18nTribute, 0)
	for _, tribute := range tributes {
		i18nTributes = append(i18nTributes, i18nTribute{
			Name:     tribute.GetItemName(),
			Emoji:    emojiService.GetItemTypeStringEmoji(tribute.GetItemType()),
			Quantity: tribute.GetQuantity() * characterNumber,
		})
	}

	return &[]*discordgo.MessageEmbed{
		{
			Title: di18n.Get(lg, "almanax.resource.title", di18n.Vars{
				"startDate": startDate.Unix(),
				"endDate":   endDate.Unix(),
			}),
			Description: di18n.Get(lg, "almanax.resource.description", di18n.Vars{
				"number":   characterNumber,
				"tributes": i18nTributes,
			}),
			Color:     constants.Color,
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: constants.GetUnknownSeason().AlmanaxIcon},
			Author: &discordgo.MessageEmbedAuthor{
				Name:    almanaxResources.GetSource().GetName(),
				URL:     almanaxResources.GetSource().GetUrl(),
				IconURL: almanaxResources.GetSource().GetIcon(),
			},
			Footer: discord.BuildDefaultFooter(lg),
		},
	}
}

func mapAlmanaxResourceToComponents(duration, characterNumber int64,
	lg discordgo.Locale) *[]discordgo.MessageComponent {
	durationCustomID := contract.CraftAlmanaxResourceDurationCustomID(characterNumber)
	durations := constants.GetAlmanaxDayDuration()
	// The command accepts durations the menu does not propose: highlight the
	// closest proposed one instead of leaving the menu blank.
	selectedDuration := nearestNumber(durations, duration)
	durationValues := make([]discordgo.SelectMenuOption, 0)
	for _, number := range durations {
		durationValues = append(durationValues, discordgo.SelectMenuOption{
			Label: di18n.Get(lg, "almanax.resource.duration.label", di18n.Vars{
				"number": number,
			}),
			Value:   fmt.Sprintf("%v", number),
			Default: selectedDuration == number,
		})
	}

	characterCustomID := contract.CraftAlmanaxResourceCharacterCustomID(duration)
	characterNumbers := make([]discordgo.SelectMenuOption, 0)
	for _, number := range constants.GetCharacterNumbers() {
		characterNumbers = append(characterNumbers, discordgo.SelectMenuOption{
			Label: di18n.Get(lg, "almanax.resource.character.label", di18n.Vars{
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
					Placeholder: di18n.Get(lg, "almanax.resource.duration.placeholder"),
					Options:     durationValues,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    characterCustomID,
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: di18n.Get(lg, "almanax.resource.character.placeholder"),
					Options:     characterNumbers,
				},
			},
		},
	}
}

// nearestNumber returns the value of numbers closest to wanted, favouring the
// smallest one on a tie; wanted itself when numbers is empty.
func nearestNumber(numbers []int64, wanted int64) int64 {
	nearest := wanted
	var bestGap int64
	for i, number := range numbers {
		gap := max(number-wanted, wanted-number)
		if i == 0 || gap < bestGap {
			nearest, bestGap = number, gap
		}
	}

	return nearest
}
