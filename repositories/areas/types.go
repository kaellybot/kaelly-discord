package areas

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetAreas() ([]entities.Area, error)
}

type Impl struct {
	db databases.MySQLConnection
}
