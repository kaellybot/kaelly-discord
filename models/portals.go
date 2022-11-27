package models

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func MapPortalPositionRequest(dimension Dimension, server Server, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_PORTAL_POSITION_REQUEST,
		Language: MapDiscordLocale(lg),
		PortalPositionRequest: &amqp.PortalPositionRequest{
			Dimension: dimension.Name,
			Server:    server.Name,
		},
	}
}

func MapToEmbed(portal amqp.PortalPositionAnswer_PortalPosition, lg discordgo.Locale) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: portal.Dimension,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    portal.Source.Name,
			IconURL: portal.Source.Icon,
			URL:     portal.Source.Url,
		},
		//TODO color, dimension name, etc.
		Footer: &discordgo.MessageEmbedFooter{
			Text: i18n.Get(lg, "pos.footer", i18n.Vars{
				"createdBy": portal.CreatedBy, "createdAt": portal.CreatedAt,
				"updatedBy": portal.UpdatedBy, "updatedAt": portal.UpdatedAt,
			}),
		},
	}
}
