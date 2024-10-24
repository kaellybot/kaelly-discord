package subareas

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetSubAreas() ([]entities.SubArea, error) {
	var subAreas []entities.SubArea
	response := repo.db.GetDB().
		Model(&entities.SubArea{}).
		Preload("Labels").
		Find(&subAreas)
	return subAreas, response.Error
}
