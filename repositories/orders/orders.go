package orders

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/spf13/viper"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetOrders() ([]entities.Order, error) {
	var orders []entities.Order
	response := repo.db.GetDB().
		Model(&entities.Order{}).
		Where("game = ?", constants.GetGame().AMQPGame).
		Preload("Labels")

	if !viper.GetBool(constants.Production) {
		response = response.
			Select("id, emoji_dark_dev AS emoji_dark, emoji_light_dev AS emoji_light, game")
	}

	response = response.Find(&orders)
	return orders, response.Error
}
