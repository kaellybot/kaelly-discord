package portals

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
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

func New(dimensionRepo dimensions.Repository, areaRepo areas.Repository,
	subAreaRepo subareas.Repository, transportTypeRepo transports.Repository) (*Impl, error) {
	// dimensions
	dimEntities, err := dimensionRepo.GetDimensions()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(dimEntities)).
		Msgf("Dimensions loaded")

	dimensions := make(map[string]entities.Dimension)
	for _, dimension := range dimEntities {
		dimensions[dimension.ID] = dimension
	}

	// area
	areaEntities, err := areaRepo.GetAreas()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(areaEntities)).
		Msgf("Areas loaded")

	areas := make(map[string]entities.Area)
	for _, area := range areaEntities {
		areas[area.ID] = area
	}

	// sub area
	subAreaEntities, err := subAreaRepo.GetSubAreas()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(subAreaEntities)).
		Msgf("Sub Areas loaded")

	subAreas := make(map[string]entities.SubArea)
	for _, subArea := range subAreaEntities {
		subAreas[subArea.ID] = subArea
	}

	// transport type
	transportTypeEntities, err := transportTypeRepo.GetTransportTypes()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(transportTypeEntities)).
		Msgf("Transport types loaded")

	transportTypes := make(map[string]entities.TransportType)
	for _, transportType := range transportTypeEntities {
		transportTypes[transportType.ID] = transportType
	}

	return &Impl{
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

func (service *Impl) GetDimension(id string) (entities.Dimension, bool) {
	dimension, found := service.dimensions[id]
	return dimension, found
}

func (service *Impl) FindDimensions(name string, locale discordgo.Locale, limit int) []entities.Dimension {
	dimensionsFound := make([]entities.Dimension, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize dimension name, returning empty dimension list")
		return dimensionsFound
	}

	for _, dimension := range service.dimensions {
		currentCleanedName, _, errStr := transform.String(service.transformer,
			strings.ToLower(translators.GetEntityLabel(dimension, locale)))
		if errStr == nil && strings.Contains(currentCleanedName, cleanedName) {
			if currentCleanedName == cleanedName {
				return []entities.Dimension{dimension}
			}

			dimensionsFound = append(dimensionsFound, dimension)
		}
	}

	if len(dimensionsFound) > limit {
		return dimensionsFound[:limit]
	}

	return dimensionsFound
}

func (service *Impl) GetArea(id string) (entities.Area, bool) {
	area, found := service.areas[id]
	return area, found
}

func (service *Impl) GetSubArea(id string) (entities.SubArea, bool) {
	subArea, found := service.subAreas[id]
	return subArea, found
}

func (service *Impl) GetTransportType(id string) (entities.TransportType, bool) {
	transportType, found := service.transportTypes[id]
	return transportType, found
}
