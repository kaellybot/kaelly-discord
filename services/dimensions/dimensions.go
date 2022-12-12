package dimensions

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/dimensions"
	"github.com/kaellybot/kaelly-discord/utils/i18n"
)

type DimensionService interface {
	GetDimension(id string) (entities.Dimension, bool)
	GetDimensions() []entities.Dimension
	FindDimensions(name string, locale discordgo.Locale) []entities.Dimension
}

type DimensionServiceImpl struct {
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
	cleanedName := strings.ToLower(name)

	// TODO normalize names

	for _, dimension := range service.dimensions {
		if strings.HasPrefix(strings.ToLower(i18n.GetEntityLabel(dimension, locale)), cleanedName) {
			dimensionsFound = append(dimensionsFound, dimension)
		}
	}

	return dimensionsFound
}
