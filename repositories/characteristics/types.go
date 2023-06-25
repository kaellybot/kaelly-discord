package cities

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetCharacteristics() ([]entities.Characteristic, error)
}

type Impl struct {
	db databases.MySQLConnection
}
