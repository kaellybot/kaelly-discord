package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	"github.com/rs/zerolog/log"
)

const (
	CommandNameAbout = "about"
)

func About(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "about.title",
					Description: "about.desc",
					Color:       models.Color,
					Image:       &discordgo.MessageEmbedImage{URL: models.AvatarImage},
					Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: models.AvatarIcon},
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "about.invite.title",
							Value:  "about.invite.desc",
							Inline: true,
						},
						{
							Name:   "about.support.title",
							Value:  "about.support.desc",
							Inline: true,
						},
						{
							Name:   "about.twitter.title",
							Value:  "about.twitter.desc",
							Inline: true,
						},
						{
							Name:   "about.opensource.title",
							Value:  "about.opensource.desc",
							Inline: true,
						},
						{
							Name:   "about.free.title",
							Value:  "about.free.desc",
							Inline: true,
						},
						{
							Name:   "about.graphist.title",
							Value:  "about.graphist.desc",
							Inline: true,
						},
						{
							Name:   "about.donators.title",
							Value:  "about.donators.desc",
							Inline: true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle about reponse")
	}
}
