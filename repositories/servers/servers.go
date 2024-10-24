package servers

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/spf13/viper"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetServers() ([]entities.Server, error) {
	var servers []entities.Server
	response := repo.db.GetDB().
		Model(&entities.Server{}).
		Where("game = ?", constants.GetGame().AMQPGame).
		Preload("Labels")

	if !viper.GetBool(constants.Production) {
		response = response.
			Select("id, icon, emoji_dev AS emoji, game")
	}

	response = response.Find(&servers)
	return servers, response.Error
}
