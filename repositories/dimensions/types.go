package dimensions

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetDimensions() ([]entities.Dimension, error)
}

type Impl struct {
	db databases.MySQLConnection
}
