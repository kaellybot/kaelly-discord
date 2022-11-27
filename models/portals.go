package models

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
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

func MapPortalsToEmbeds() *[]*discordgo.MessageEmbed {
	// TODO
	return nil
}
