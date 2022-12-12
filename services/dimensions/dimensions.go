package dimensions

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/dimensions"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type DimensionService interface {
	GetDimension(id string) (entities.Dimension, bool)
	GetDimensions() []entities.Dimension
	FindDimensions(name string, locale discordgo.Locale) []entities.Dimension
}

type DimensionServiceImpl struct {
	transformer   transform.Transformer
	dimensionsMap map[string]entities.Dimension
	dimensions    []entities.Dimension
	repository    repository.DimensionRepository
}

func New(repository repository.DimensionRepository) (*DimensionServiceImpl, error) {
	dimensions, err := repository.GetDimensions()
	if err != nil {
		return nil, err
	}

	dimensionsMap := make(map[string]entities.Dimension)
	for _, dimension := range dimensions {
		dimensionsMap[dimension.Id] = dimension
	}

	return &DimensionServiceImpl{
		transformer:   transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		dimensionsMap: dimensionsMap,
		dimensions:    dimensions,
		repository:    repository,
	}, nil
}

func (service *DimensionServiceImpl) GetDimensions() []entities.Dimension {
	return service.dimensions
}

func (service *DimensionServiceImpl) GetDimension(id string) (entities.Dimension, bool) {
	dimension, found := service.dimensionsMap[id]
	return dimension, found
}

func (service *DimensionServiceImpl) FindDimensions(name string, locale discordgo.Locale) []entities.Dimension {
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
