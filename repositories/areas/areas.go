package areas

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type AreaRepository interface {
	GetAreas() ([]entities.Area, error)
}

type AreaRepositoryImpl struct {
	db databases.MySQLConnection
}

func New(db databases.MySQLConnection) *AreaRepositoryImpl {
	return &AreaRepositoryImpl{db: db}
}

func (repo *AreaRepositoryImpl) GetAreas() ([]entities.Area, error) {
	var Areas []entities.Area
	response := repo.db.GetDB().Model(&entities.Area{}).Preload("Labels").Find(&Areas)
	return Areas, response.Error
}
