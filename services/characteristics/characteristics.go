package characteristics

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/characteristics"
	"github.com/rs/zerolog/log"
)

func New(repository repository.Repository) (*Impl, error) {
	characteristics, err := repository.GetCharacteristics()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(characteristics)).
		Msgf("Characteristics loaded")

	characteristicsMap := make(map[string]entities.Characteristic)
	for _, characteristic := range characteristics {
		characteristicsMap[characteristic.ID] = characteristic
	}

	return &Impl{
		characteristics: characteristicsMap,
		repository:      repository,
	}, nil
}

func (service *Impl) GetCharacteristic(id string) (entities.Characteristic, bool) {
	characteristic, found := service.characteristics[id]
	return characteristic, found
}
