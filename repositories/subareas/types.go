package subareas

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetSubAreas() ([]entities.SubArea, error)
}

type Impl struct {
	db databases.MySQLConnection
}
