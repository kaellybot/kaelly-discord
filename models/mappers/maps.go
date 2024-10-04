package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func MapCompetitionMapRequest(mapNumber int64, lg discordgo.Locale,
) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_COMPETITION_MAP_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		Game:     constants.GetGame().AMQPGame,
		CompetitionMapRequest: &amqp.CompetitionMapRequest{
			MapNumber: mapNumber,
		},
	}
}

func MapCompetitionMapToWebhookEdit(answer *amqp.CompetitionMapAnswer, mapType constants.MapType,
	service emojis.Service, locale amqp.Language) *discordgo.WebhookEdit {
	lg := constants.MapAMQPLocale(locale)
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
		Title: i18n.Get(lg, "map.title", i18n.Vars{
			"mapNumber": competitiveMap.MapNumber,
		}),
		Description: i18n.Get(lg, "map.taunt"),
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
			Label:    i18n.Get(lg, "map.button.tactical"),
			Style:    discordgo.PrimaryButton,
			Emoji:    service.GetMiscEmoji(constants.EmojiIDTacticalMap),
		})
	case constants.MapTypeTactical:
		components = append(components, discordgo.Button{
			CustomID: contract.CraftMapNormalCustomID(competitiveMap.GetMapNumber()),
			Label:    i18n.Get(lg, "map.button.normal"),
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
