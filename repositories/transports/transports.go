package transports

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/spf13/viper"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetTransportTypes() ([]entities.TransportType, error) {
	var transportTypes []entities.TransportType
	response := repo.db.GetDB().
		Model(&entities.TransportType{}).
		Preload("Labels")

	if !viper.GetBool(constants.Production) {
		response = response.
			Select("id, emoji_dev AS emoji")
	}

	response = response.Find(&transportTypes)
	return transportTypes, response.Error
}
