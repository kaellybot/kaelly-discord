package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	di18n "github.com/kaysoro/discordgo-i18n"
)

func MapCompetitionMapRequest(mapNumber int64, authorID string, lg discordgo.Locale,
) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_COMPETITION_MAP_REQUEST, lg)
	request.CompetitionMapRequest = &amqp.CompetitionMapRequest{
		MapNumber: mapNumber,
	}
	return request
}

func MapCompetitionMapToWebhookEdit(answer *amqp.CompetitionMapAnswer, mapType constants.MapType,
	service emojis.Service, locale amqp.Language) *discordgo.WebhookEdit {
	lg := i18n.MapAMQPLocale(locale)
	return &discordgo.WebhookEdit{
		Embeds:     mapCompetitionMapToEmbed(answer, mapType, lg),
		Components: mapCompetitionMapToComponents(answer, mapType, service, lg),
	}
}

func mapCompetitionMapToEmbed(competitiveMap *amqp.CompetitionMapAnswer,
	mapType constants.MapType, lg discordgo.Locale) *[]*discordgo.MessageEmbed {
	var imageURL string
	switch mapType {
	case constants.MapTypeNormal:
		imageURL = competitiveMap.MapNormalURL
	case constants.MapTypeTactical:
		imageURL = competitiveMap.MapTacticalURL
	default:
		imageURL = ""
	}

	embed := discordgo.MessageEmbed{
		Title: di18n.Get(lg, "map.title", di18n.Vars{
			"mapNumber": competitiveMap.MapNumber,
		}),
		Description: di18n.Get(lg, "map.taunt"),
		Color:       constants.Color,
		Image:       &discordgo.MessageEmbedImage{URL: imageURL},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    competitiveMap.Source.Name,
			URL:     competitiveMap.Source.Url,
			IconURL: competitiveMap.Source.Icon,
		},
		Footer: discord.BuildDefaultFooter(lg),
	}

	return &[]*discordgo.MessageEmbed{&embed}
}

func mapCompetitionMapToComponents(competitiveMap *amqp.CompetitionMapAnswer, mapType constants.MapType,
	service emojis.Service, lg discordgo.Locale) *[]discordgo.MessageComponent {
	components := make([]discordgo.MessageComponent, 0)
	switch mapType {
	case constants.MapTypeNormal:
		components = append(components, discordgo.Button{
			CustomID: contract.CraftMapTacticalCustomID(competitiveMap.GetMapNumber()),
			Label:    di18n.Get(lg, "map.button.tactical"),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDTacticalMap),
		})
	case constants.MapTypeTactical:
		components = append(components, discordgo.Button{
			CustomID: contract.CraftMapNormalCustomID(competitiveMap.GetMapNumber()),
			Label:    di18n.Get(lg, "map.button.normal"),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDNormalMap),
		})
	}

	return &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: components,
		},
	}
}
