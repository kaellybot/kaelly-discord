package portals

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/areas"
	"github.com/kaellybot/kaelly-discord/repositories/dimensions"
	"github.com/kaellybot/kaelly-discord/repositories/subareas"
	"github.com/kaellybot/kaelly-discord/repositories/transports"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type PortalService interface {
	GetDimension(id string) (entities.Dimension, bool)
	GetArea(id string) (entities.Area, bool)
	GetSubArea(id string) (entities.SubArea, bool)
	GetTransportType(id string) (entities.TransportType, bool)
	FindDimensions(name string, locale discordgo.Locale) []entities.Dimension
}

type PortalServiceImpl struct {
	transformer       transform.Transformer
	dimensions        map[string]entities.Dimension
	areas             map[string]entities.Area
	subAreas          map[string]entities.SubArea
	transportTypes    map[string]entities.TransportType
	dimensionRepo     dimensions.DimensionRepository
	areaRepo          areas.AreaRepository
	subAreaRepo       subareas.SubAreaRepository
	transportTypeRepo transports.TransportTypeRepository
}

func New(dimensionRepo dimensions.DimensionRepository,
	areaRepo areas.AreaRepository,
	subAreaRepo subareas.SubAreaRepository,
	transportTypeRepo transports.TransportTypeRepository) (*PortalServiceImpl, error) {

	// dimensions
	dimEntities, err := dimensionRepo.GetDimensions()
	if err != nil {
		return nil, err
	}

	dimensions := make(map[string]entities.Dimension)
	for _, dimension := range dimEntities {
		dimensions[dimension.Id] = dimension
	}

	// area
	areaEntities, err := areaRepo.GetAreas()
	if err != nil {
		return nil, err
	}

	areas := make(map[string]entities.Area)
	for _, area := range areaEntities {
		areas[area.Id] = area
	}

	// sub area
	subAreaEntities, err := subAreaRepo.GetSubAreas()
	if err != nil {
		return nil, err
	}

	subAreas := make(map[string]entities.SubArea)
	for _, subArea := range subAreaEntities {
		subAreas[subArea.Id] = subArea
	}

	// transport type
	transportTypeEntities, err := transportTypeRepo.GetTransportTypes()
	if err != nil {
		return nil, err
	}

	transportTypes := make(map[string]entities.TransportType)
	for _, transportType := range transportTypeEntities {
		transportTypes[transportType.Id] = transportType
	}

	return &PortalServiceImpl{
		transformer:       transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		dimensions:        dimensions,
		dimensionRepo:     dimensionRepo,
		areas:             areas,
		areaRepo:          areaRepo,
		subAreas:          subAreas,
		subAreaRepo:       subAreaRepo,
		transportTypes:    transportTypes,
		transportTypeRepo: transportTypeRepo,
	}, nil
}

func (service *PortalServiceImpl) GetDimension(id string) (entities.Dimension, bool) {
	dimension, found := service.dimensions[id]
	return dimension, found
}

func (service *PortalServiceImpl) FindDimensions(name string, locale discordgo.Locale) []entities.Dimension {
	dimensionsFound := make([]entities.Dimension, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize dimension name, returning empty dimension list")
		return dimensionsFound
	}

	for _, dimension := range service.dimensions {
		currentCleanedName, _, err := transform.String(service.transformer, strings.ToLower(translators.GetEntityLabel(dimension, locale)))
		if err == nil && strings.HasPrefix(currentCleanedName, cleanedName) {
			dimensionsFound = append(dimensionsFound, dimension)
		}
	}

	return dimensionsFound
}

func (service *PortalServiceImpl) GetArea(id string) (entities.Area, bool) {
	area, found := service.areas[id]
	return area, found
}

func (service *PortalServiceImpl) GetSubArea(id string) (entities.SubArea, bool) {
	subArea, found := service.subAreas[id]
	return subArea, found
}

func (service *PortalServiceImpl) GetTransportType(id string) (entities.TransportType, bool) {
	transportType, found := service.transportTypes[id]
	return transportType, found
}
