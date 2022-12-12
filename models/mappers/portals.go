package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/translators"
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

func MapToEmbed(portal *amqp.PortalPositionAnswer_PortalPosition, dimension entities.Dimension, locale amqp.RabbitMQMessage_Language) *discordgo.MessageEmbed {
	lg := constants.MapAmqpLocale(locale)
	embed := discordgo.MessageEmbed{
		Title:     translators.GetEntityLabel(dimension, lg),
		Color:     dimension.Color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: dimension.Icon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    i18n.Get(lg, "pos.embed.footer", i18n.Vars{"source": portal.Source.Name}),
			IconURL: portal.Source.Icon,
		},
	}

	if portal.Position != nil {
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   i18n.Get(lg, "pos.embed.position.name"),
				Value:  i18n.Get(lg, "pos.embed.position.value", i18n.Vars{"position": portal.Position}),
				Inline: true,
			},
			{
				Name:   i18n.Get(lg, "pos.embed.uses.name"),
				Value:  i18n.Get(lg, "pos.embed.uses.value", i18n.Vars{"uses": portal.RemainingUses}),
				Inline: true,
			},
			{
				Name: i18n.Get(lg, "pos.embed.hunters.name"),
				Value: i18n.Get(lg, "pos.embed.hunters.value", i18n.Vars{
					"createdBy": portal.CreatedBy, "createdAt": portal.CreatedAt,
					"updatedBy": portal.UpdatedBy, "updatedAt": portal.UpdatedAt,
				}),
				Inline: true,
			},
		}
		//TODO zaap nearby, private
	} else {
		embed.Description = i18n.Get(lg, "pos.embed.unknown")
	}

	return &embed
}
