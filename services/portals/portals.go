package portals

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/dimensions"
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
	GetDimensions() []entities.Dimension
	FindDimensions(name string, locale discordgo.Locale) []entities.Dimension
}

type PortalServiceImpl struct {
	transformer   transform.Transformer
	dimensionsMap map[string]entities.Dimension
	dimensions    []entities.Dimension
	dimensionRepo dimensions.DimensionRepository
}

func New(dimensionRepo dimensions.DimensionRepository) (*PortalServiceImpl, error) {
	dimensions, err := dimensionRepo.GetDimensions()
	if err != nil {
		return nil, err
	}

	dimensionsMap := make(map[string]entities.Dimension)
	for _, dimension := range dimensions {
		dimensionsMap[dimension.Id] = dimension
	}

	return &PortalServiceImpl{
		transformer:   transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		dimensionsMap: dimensionsMap,
		dimensions:    dimensions,
		dimensionRepo: dimensionRepo,
	}, nil
}

func (service *PortalServiceImpl) GetDimensions() []entities.Dimension {
	return service.dimensions
}

func (service *PortalServiceImpl) GetDimension(id string) (entities.Dimension, bool) {
	dimension, found := service.dimensionsMap[id]
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
	// TODO
	return entities.Area{}, false
}

func (service *PortalServiceImpl) GetSubArea(id string) (entities.SubArea, bool) {
	// TODO
	return entities.SubArea{}, false
}

func (service *PortalServiceImpl) GetTransportType(id string) (entities.TransportType, bool) {
	// TODO
	return entities.TransportType{}, false
}
