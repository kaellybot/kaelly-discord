package cities

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetCities() ([]entities.City, error) {
	var cities []entities.City
	response := repo.db.GetDB().Model(&entities.City{}).Preload("Labels").Find(&cities)
	return cities, response.Error
}
