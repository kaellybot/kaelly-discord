package orders

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type OrderRepository interface {
	GetOrders() ([]entities.Order, error)
}

type OrderRepositoryImpl struct {
	db databases.MySQLConnection
}

func New(db databases.MySQLConnection) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{db: db}
}

func (repo *OrderRepositoryImpl) GetOrders() ([]entities.Order, error) {
	var orders []entities.Order
	response := repo.db.GetDB().Model(&entities.Order{}).Preload("Labels").Find(&orders)
	return orders, response.Error
}
