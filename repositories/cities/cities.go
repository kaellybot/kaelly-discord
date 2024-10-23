package cities

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

func (repo *Impl) GetCities() ([]entities.City, error) {
	var response *gorm.DB
	var cities []entities.City
	if viper.GetBool(constants.Production) {
		response = repo.db.GetDB().
			Model(&entities.City{}).
			Preload("Labels").
			Find(&cities)
	} else {
		response = repo.db.GetDB().
			Model(&entities.City{}).
			Select("id, icon, emoji_dev AS emoji, color, type").
			Preload("Labels").
			Find(&cities)
	}
	return cities, response.Error
}
