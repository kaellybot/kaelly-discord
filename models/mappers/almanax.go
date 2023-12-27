package mappers

import (
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapAlmanaxRequest(date *time.Time, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaAlmanaxRequest: &amqp.EncyclopediaAlmanaxRequest{
			Date: timestamppb.New(*date),
		},
	}
}

func MapAlmanaxResourceRequest(duration int32, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaAlmanaxResourceRequest: &amqp.EncyclopediaAlmanaxResourceRequest{
			Duration: duration,
		},
	}
}

func MapAlmanaxEffectListRequest(query string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST,
		Language: constants.MapDiscordLocale(lg),
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
		EncyclopediaAlmanaxEffectRequest: &amqp.EncyclopediaAlmanaxEffectRequest{
			Query: query,
		},
	}
}

func MapAlmanaxToEmbed(almanax *amqp.Almanax, lg discordgo.Locale) *discordgo.MessageEmbed {
	season := constants.GetSeason(almanax.Date.AsTime())
	return &discordgo.MessageEmbed{
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
					"quantity": almanax.Tribute.Quantity,
				}),
			},
		},
	}
}

func MapAlmanaxResourceToEmbed(almanaxResources *amqp.EncyclopediaAlmanaxResourceAnswer,
	locale amqp.Language) *discordgo.MessageEmbed {

	lg := constants.MapAMQPLocale(locale)
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, int(almanaxResources.Duration))
	collator := constants.MapCollator(lg)
	sort.SliceStable(almanaxResources.Tributes, func(i, j int) bool {
		return collator.CompareString(almanaxResources.Tributes[i].ItemName, almanaxResources.Tributes[j].ItemName) == -1
	})

	return &discordgo.MessageEmbed{
		Title: i18n.Get(lg, "almanax.resource.title", i18n.Vars{
			"startDate": startDate.Unix(),
			"endDate":   endDate.Unix(),
		}),
		Description: i18n.Get(lg, "almanax.resource.description", i18n.Vars{"tributes": almanaxResources.Tributes}),
		Color:       constants.Color,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: constants.GetUnknownSeason().AlmanaxIcon},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    almanaxResources.Source.Name,
			URL:     almanaxResources.Source.Url,
			IconURL: almanaxResources.Source.Icon,
		},
		Footer: discord.BuildDefaultFooter(lg),
	}
}
