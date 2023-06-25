package characteristics

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/characteristics"
)

type Service interface {
	GetCharacteristic(id string) (entities.Characteristic, bool)
}

type Impl struct {
	characteristics map[string]entities.Characteristic
	repository      repository.Repository
}
