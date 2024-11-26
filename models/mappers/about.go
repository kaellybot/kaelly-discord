package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func MapAboutRequest(authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return requestBackbone(authorID, amqp.RabbitMQMessage_ABOUT_REQUEST, lg)
}

func MapAboutToWebhook(lg discordgo.Locale, emojiService emojis.Service) *discordgo.WebhookEdit {
	return &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title: i18n.Get(lg, "about.title", i18n.Vars{"name": constants.Name, "version": constants.Version}),
				Description: i18n.Get(lg, "about.desc", i18n.Vars{
					"game":     constants.GetGame(),
					"gameLogo": emojiService.GetMiscStringEmoji(constants.EmojiIDGame),
				}),
				Color:     constants.Color,
				Image:     &discordgo.MessageEmbedImage{URL: constants.AvatarImage},
				Thumbnail: &discordgo.MessageEmbedThumbnail{URL: constants.GetGame().Icon},
				Footer: &discordgo.MessageEmbedFooter{
					Text:    i18n.Get(lg, "about.footer"),
					IconURL: constants.AnkamaLogo,
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   i18n.Get(lg, "about.support.title"),
						Value:  i18n.Get(lg, "about.support.desc", i18n.Vars{"discord": constants.Discord}),
						Inline: false,
					},
					{
						Name: i18n.Get(lg, "about.twitter.title", i18n.Vars{
							"twitterLogo": emojiService.GetMiscStringEmoji(constants.EmojiIDTwitter),
						}),
						Value:  i18n.Get(lg, "about.twitter.desc", i18n.Vars{"twitter": constants.Twitter}),
						Inline: false,
					},
					{
						Name: i18n.Get(lg, "about.opensource.title", i18n.Vars{
							"githubLogo": emojiService.GetMiscStringEmoji(constants.EmojiIDGithub),
						}),
						Value:  i18n.Get(lg, "about.opensource.desc", i18n.Vars{"github": constants.Github}),
						Inline: false,
					},
					{
						Name:   i18n.Get(lg, "about.free.title"),
						Value:  i18n.Get(lg, "about.free.desc", i18n.Vars{"paypal": constants.Paypal}),
						Inline: false,
					},
					{
						Name:   i18n.Get(lg, "about.privacy.title"),
						Value:  i18n.Get(lg, "about.privacy.desc"),
						Inline: false,
					},
					{
						Name: i18n.Get(lg, "about.graphist.title"),
						Value: i18n.Get(lg, "about.graphist.desc", i18n.Vars{
							"Elycann": constants.GetGraphistElycann(),
							"Colibry": constants.GetGraphistColibry(),
						}),
						Inline: false,
					},
				},
			},
		},
	}
}
