package almanaxes

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetAlmanaxNews() ([]entities.AlmanaxNews, error)
}

type Impl struct {
	db databases.MySQLConnection
}
