package characteristics

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetCharacteristics() ([]entities.Characteristic, error) {
	var characteristics []entities.Characteristic
	response := repo.db.GetDB().
		Model(entities.Characteristic{})

	if !viper.GetBool(constants.Production) {
		response = response.
			Select("id, emoji_dev AS emoji, sort_order").
			Preload("Regexes", func(db *gorm.DB) *gorm.DB {
				return db.Select("characteristic_id, expression, emoji_dev AS emoji")
			})
	} else {
		response = response.
			Preload("Regexes")
	}

	response = response.Find(&characteristics)
	return characteristics, response.Error
}
