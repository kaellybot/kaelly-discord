package dimensions

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
)

type DimensionService interface {
	GetDimensions() []entities.Dimension
	FindDimensions(name string, locale discordgo.Locale) []entities.Dimension
}

type DimensionServiceImpl struct {
	dimensions []entities.Dimension
}

func New() (*DimensionServiceImpl, error) {
	return &DimensionServiceImpl{
		dimensions: []entities.Dimension{
			{Id: "Enutrosor"},
			{Id: "Srambad"},
			{Id: "Xelorium"},
			{Id: "Ecaflipus"},
		},
	}, nil
}

func (service *DimensionServiceImpl) GetDimensions() []entities.Dimension {
	return service.dimensions
}

func (service *DimensionServiceImpl) FindDimensions(name string, locale discordgo.Locale) []entities.Dimension {
	dimensionsFound := make([]entities.Dimension, 0)
	cleanedName := strings.ToLower(name)

	for _, dimension := range service.dimensions {
		if strings.HasPrefix(strings.ToLower(dimension.Id), cleanedName) {
			dimensionsFound = append(dimensionsFound, dimension)
		}
	}

	return dimensionsFound
}
