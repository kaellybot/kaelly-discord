package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func MapPortalPositionRequest(dimension entities.Dimension, server entities.Server,
	lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_PORTAL_POSITION_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		PortalPositionRequest: &amqp.PortalPositionRequest{
			DimensionId: dimension.ID,
			ServerId:    server.ID,
		},
	}
}

func MapPortalToEmbed(portal *amqp.PortalPositionAnswer_PortalPosition, portalService portals.Service,
	serverService servers.Service, locale amqp.Language) *discordgo.MessageEmbed {
	lg := constants.MapAMQPLocale(locale)
	dimension, found := portalService.GetDimension(portal.DimensionId)
	if !found {
		log.Warn().Str(constants.LogEntity, portal.DimensionId).
			Msgf("Cannot find dimension based on ID sent internally, continuing with empty dimension")
		dimension = entities.Dimension{ID: portal.DimensionId}
	}

	server, found := serverService.GetServer(portal.ServerId)
	if !found {
		log.Warn().Str(constants.LogEntity, portal.ServerId).
			Msgf("Cannot find server based on ID sent internally, continuing with empty server")
		server = entities.Server{ID: portal.ServerId}
	}

	embed := discordgo.MessageEmbed{
		Title:     translators.GetEntityLabel(dimension, lg),
		Color:     dimension.Color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: dimension.Icon},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    portal.Source.Name,
			URL:     portal.Source.Url,
			IconURL: portal.Source.Icon,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    translators.GetEntityLabel(server, lg),
			IconURL: server.Icon,
		},
	}

	if portal.Position != nil {
		embed.Description = i18n.Get(lg, "pos.embed.known", i18n.Vars{
			"position":  portal.Position,
			"uses":      portal.RemainingUses,
			"createdBy": portal.CreatedBy, "createdAt": portal.CreatedAt,
			"updatedBy": portal.UpdatedBy, "updatedAt": portal.UpdatedAt,
		})

		if portal.Position.ConditionalTransport != nil {
			embed.Fields = append(embed.Fields,
				mapTransportToEmbed(portal.Position.ConditionalTransport, portalService, lg))
		}

		embed.Fields = append(embed.Fields,
			mapTransportToEmbed(portal.Position.Transport, portalService, lg))
	} else {
		embed.Description = i18n.Get(lg, "pos.embed.unknown")
	}

	return &embed
}

func mapTransportToEmbed(transport *amqp.PortalPositionAnswer_PortalPosition_Position_Transport,
	portalService portals.Service, lg discordgo.Locale) *discordgo.MessageEmbedField {
	transportType, found := portalService.GetTransportType(transport.TypeId)
	if !found {
		log.Warn().
			Str(constants.LogEntity, transport.TypeId).
			Msgf("Cannot find transport type based on ID sent internally, continuing with empty transport")
		transportType = entities.TransportType{ID: transport.TypeId}
	}

	area, found := portalService.GetArea(transport.AreaId)
	if !found {
		log.Warn().
			Str(constants.LogEntity, transport.AreaId).
			Msgf("Cannot find area based on ID sent internally, continuing with empty area")
		area = entities.Area{ID: transport.AreaId}
	}

	subArea, found := portalService.GetSubArea(transport.SubAreaId)
	if !found {
		log.Warn().
			Str(constants.LogEntity, transport.SubAreaId).
			Msgf("Cannot find sub area based on ID sent internally, continuing with empty sub area")
		subArea = entities.SubArea{ID: transport.SubAreaId}
	}

	return &discordgo.MessageEmbedField{
		Name: i18n.Get(lg, "pos.embed.transport.name", i18n.Vars{
			"type":  translators.GetEntityLabel(transportType, lg),
			"emoji": transportType.Emoji,
		}),
		Value: i18n.Get(lg, "pos.embed.transport.value", i18n.Vars{
			"area":    translators.GetEntityLabel(area, lg),
			"subArea": translators.GetEntityLabel(subArea, lg),
			"x":       transport.X,
			"y":       transport.Y,
		}),
		Inline: true,
	}
}
