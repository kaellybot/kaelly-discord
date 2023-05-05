package areas

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetAreas() ([]entities.Area, error) {
	var Areas []entities.Area
	response := repo.db.GetDB().Model(&entities.Area{}).Preload("Labels").Find(&Areas)
	return Areas, response.Error
}
