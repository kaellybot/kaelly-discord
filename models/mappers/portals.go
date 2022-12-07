package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func MapPortalPositionRequest(dimension entities.Dimension, server entities.Server, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_PORTAL_POSITION_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		PortalPositionRequest: &amqp.PortalPositionRequest{
			Dimension: dimension.Id,
			Server:    server.Id,
		},
	}
}

func MapToEmbed(portal *amqp.PortalPositionAnswer_PortalPosition, locale amqp.RabbitMQMessage_Language) *discordgo.MessageEmbed {
	lg := constants.MapAmqpLocale(locale)
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
