package cities

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type CityRepository interface {
	GetCities() ([]entities.City, error)
}

type CityRepositoryImpl struct {
	db databases.MySQLConnection
}

func New(db databases.MySQLConnection) *CityRepositoryImpl {
	return &CityRepositoryImpl{db: db}
}

func (repo *CityRepositoryImpl) GetCities() ([]entities.City, error) {
	var cities []entities.City
	response := repo.db.GetDB().Model(&entities.City{}).Preload("Labels").Find(&cities)
	return cities, response.Error
}
