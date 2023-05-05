package orders

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetOrders() ([]entities.Order, error)
}

type Impl struct {
	db databases.MySQLConnection
}
