package orders

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

func (repo *Impl) GetOrders() ([]entities.Order, error) {
	var response *gorm.DB
	var orders []entities.Order
	if viper.GetBool(constants.Production) {
		response = repo.db.GetDB().
			Model(&entities.Order{}).
			Preload("Labels").
			Find(&orders)
	} else {
		response = repo.db.GetDB().
			Model(&entities.Order{}).
			Select("id, emoji_dark_dev AS emoji_dark, emoji_light_dev AS emoji_light").
			Preload("Labels").
			Find(&orders)
	}
	return orders, response.Error
}
