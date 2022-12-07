package dimensions

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/dimensions"
)

type DimensionService interface {
	GetDimensions() []entities.Dimension
	FindDimensions(name string, locale discordgo.Locale) []entities.Dimension
}

type DimensionServiceImpl struct {
	dimensions []entities.Dimension
	repository repository.DimensionRepository
}

func New(repository repository.DimensionRepository) (*DimensionServiceImpl, error) {
	dimensions, err := repository.GetDimensions()
	if err != nil {
		return nil, err
	}
	return &DimensionServiceImpl{
		dimensions: dimensions,
		repository: repository,
	}, nil
}

func (service *DimensionServiceImpl) GetDimensions() []entities.Dimension {
	return service.dimensions
}

func (service *DimensionServiceImpl) FindDimensions(name string, locale discordgo.Locale) []entities.Dimension {
	dimensionsFound := make([]entities.Dimension, 0)
	cleanedName := strings.ToLower(name)

	// TODO not based on id

	for _, dimension := range service.dimensions {
		if strings.HasPrefix(strings.ToLower(dimension.Id), cleanedName) {
			dimensionsFound = append(dimensionsFound, dimension)
		}
	}

	return dimensionsFound
}
