package transports

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

func (repo *Impl) GetTransportTypes() ([]entities.TransportType, error) {
	var response *gorm.DB
	var transportTypes []entities.TransportType
	if viper.GetBool(constants.Production) {
		response = repo.db.GetDB().
			Model(&entities.TransportType{}).
			Preload("Labels").
			Find(&transportTypes)
	} else {
		response = repo.db.GetDB().
			Model(&entities.TransportType{}).
			Select("id, emoji_dev AS emoji").
			Preload("Labels").
			Find(&transportTypes)
	}

	return transportTypes, response.Error
}
