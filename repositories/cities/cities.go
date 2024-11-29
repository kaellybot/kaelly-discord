package cities

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/spf13/viper"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetCities() ([]entities.City, error) {
	var cities []entities.City
	response := repo.db.GetDB().
		Model(&entities.City{}).
		Where("game = ?", constants.GetGame().AMQPGame).
		Preload("Labels")

	if !viper.GetBool(constants.Production) {
		response = response.
			Select("id, icon, emoji_dev AS emoji, color, type, game")
	}

	response = response.Find(&cities)
	return cities, response.Error
}
