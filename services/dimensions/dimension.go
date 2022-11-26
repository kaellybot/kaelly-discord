package dimensions

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
)

type DimensionService interface {
	GetDimensions() []models.Dimension
	FindDimensions(name string, locale discordgo.Locale) []models.Dimension
}

type DimensionServiceImpl struct {
	dimensions []models.Dimension
}

func New() (*DimensionServiceImpl, error) {
	return &DimensionServiceImpl{
		dimensions: []models.Dimension{
			{Name: "Enutrosor"},
			{Name: "Srambad"},
			{Name: "Xelorium"},
			{Name: "Ecaflipus"},
		},
	}, nil
}

func (service *DimensionServiceImpl) GetDimensions() []models.Dimension {
	return service.dimensions
}

func (service *DimensionServiceImpl) FindDimensions(name string, locale discordgo.Locale) []models.Dimension {
	dimensionsFound := make([]models.Dimension, 0)
	cleanedName := strings.ToLower(name)

	for _, dimension := range service.dimensions {
		if strings.HasPrefix(strings.ToLower(dimension.Name), cleanedName) {
			dimensionsFound = append(dimensionsFound, dimension)
		}
	}

	return dimensionsFound
}
