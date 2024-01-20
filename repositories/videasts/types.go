package videasts

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetVideasts() ([]entities.Videast, error)
}

type Impl struct {
	db databases.MySQLConnection
}
