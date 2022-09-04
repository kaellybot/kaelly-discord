package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

const (
	CommandNameAbout = "about"
)

func About(s *discordgo.Session, i *discordgo.InteractionCreate) {

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{getAboutEmbed(i.Locale)},
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle about reponse")
	}
}

func getAboutEmbed(locale discordgo.Locale) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       i18n.Get(locale, "about.title", map[string]interface{}{"name": models.Name, "version": models.Version}),
		Description: i18n.Get(locale, "about.desc"),
		Color:       models.Color,
		Image:       &discordgo.MessageEmbedImage{URL: models.AvatarImage},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: models.AvatarIcon},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   i18n.Get(locale, "about.invite.title"),
				Value:  i18n.Get(locale, "about.invite.desc"),
				Inline: true,
			},
			{
				Name:   i18n.Get(locale, "about.support.title"),
				Value:  i18n.Get(locale, "about.support.desc"),
				Inline: true,
			},
			{
				Name:   i18n.Get(locale, "about.twitter.title"),
				Value:  i18n.Get(locale, "about.twitter.desc"),
				Inline: true,
			},
			{
				Name:   i18n.Get(locale, "about.opensource.title"),
				Value:  i18n.Get(locale, "about.opensource.desc"),
				Inline: true,
			},
			{
				Name:   i18n.Get(locale, "about.free.title"),
				Value:  i18n.Get(locale, "about.free.desc"),
				Inline: true,
			},
			{
				Name:   i18n.Get(locale, "about.graphist.title"),
				Value:  i18n.Get(locale, "about.graphist.desc"),
				Inline: true,
			},
			{
				Name:   i18n.Get(locale, "about.donators.title"),
				Value:  i18n.Get(locale, "about.donators.desc"),
				Inline: true,
			},
		},
	}
}
