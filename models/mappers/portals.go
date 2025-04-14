package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	di18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func MapPortalPositionRequest(dimension entities.Dimension, server entities.Server,
	authorID string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	request := requestBackbone(authorID, amqp.RabbitMQMessage_PORTAL_POSITION_REQUEST, lg)
	request.PortalPositionRequest = &amqp.PortalPositionRequest{
		DimensionId: dimension.ID,
		ServerId:    server.ID,
	}
	return request
}

func MapPortalToEmbed(portal *amqp.PortalPositionAnswer_PortalPosition, portalService portals.Service,
	serverService servers.Service, emojiService emojis.Service, locale amqp.Language,
) *discordgo.MessageEmbed {
	lg := i18n.MapAMQPLocale(locale)
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
		embed.Description = di18n.Get(lg, "pos.embed.known", di18n.Vars{
			"position":  portal.Position,
			"uses":      portal.RemainingUses,
			"createdBy": portal.CreatedBy, "createdAt": portal.CreatedAt,
			"updatedBy": portal.UpdatedBy, "updatedAt": portal.UpdatedAt,
		})

		if portal.Position.ConditionalTransport != nil {
			embed.Fields = append(embed.Fields,
				mapTransportToEmbed(portal.Position.ConditionalTransport, portalService, emojiService, lg))
		}

		embed.Fields = append(embed.Fields,
			mapTransportToEmbed(portal.Position.Transport, portalService, emojiService, lg))
	} else {
		embed.Description = di18n.Get(lg, "pos.embed.unknown")
	}

	return &embed
}

func mapTransportToEmbed(transport *amqp.PortalPositionAnswer_PortalPosition_Position_Transport,
	portalService portals.Service, emojiService emojis.Service, lg discordgo.Locale,
) *discordgo.MessageEmbedField {
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
		Name: di18n.Get(lg, "pos.embed.transport.name", di18n.Vars{
			"type":  translators.GetEntityLabel(transportType, lg),
			"emoji": emojiService.GetEntityStringEmoji(transportType.ID, constants.EmojiTypeTransportType),
		}),
		Value: di18n.Get(lg, "pos.embed.transport.value", di18n.Vars{
			"area":    translators.GetEntityLabel(area, lg),
			"subArea": translators.GetEntityLabel(subArea, lg),
			"x":       transport.X,
			"y":       transport.Y,
		}),
		Inline: true,
	}
}
