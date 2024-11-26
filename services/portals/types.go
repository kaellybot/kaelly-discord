package portals

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/areas"
	"github.com/kaellybot/kaelly-discord/repositories/dimensions"
	"github.com/kaellybot/kaelly-discord/repositories/subareas"
	"github.com/kaellybot/kaelly-discord/repositories/transports"
	"golang.org/x/text/transform"
)

type Service interface {
	GetDimension(id string) (entities.Dimension, bool)
	GetArea(id string) (entities.Area, bool)
	GetSubArea(id string) (entities.SubArea, bool)
	GetTransportType(id string) (entities.TransportType, bool)
	FindDimensions(name string, locale discordgo.Locale, limit int) []entities.Dimension
}

type Impl struct {
	transformer       transform.Transformer
	dimensions        map[string]entities.Dimension
	areas             map[string]entities.Area
	subAreas          map[string]entities.SubArea
	transportTypes    map[string]entities.TransportType
	dimensionRepo     dimensions.Repository
	areaRepo          areas.Repository
	subAreaRepo       subareas.Repository
	transportTypeRepo transports.Repository
}
