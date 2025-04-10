package characteristics

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetCharacteristics() ([]entities.Characteristic, error) {
	var characteristics []entities.Characteristic
	response := repo.db.GetDB().
		Model(entities.Characteristic{}).
		Preload("Regexes").
		Find(&characteristics)
	return characteristics, response.Error
}
