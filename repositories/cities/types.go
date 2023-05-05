package cities

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetCities() ([]entities.City, error)
}

type Impl struct {
	db databases.MySQLConnection
}
