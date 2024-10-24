package dimensions

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetDimensions() ([]entities.Dimension, error) {
	var dimensions []entities.Dimension
	response := repo.db.GetDB().
		Model(&entities.Dimension{}).
		Preload("Labels").
		Find(&dimensions)
	return dimensions, response.Error
}
