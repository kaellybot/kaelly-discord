package servers

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetServers() ([]entities.Server, error) {
	var servers []entities.Server
	response := repo.db.GetDB().Model(&entities.Server{}).Preload("Labels").Find(&servers)
	return servers, response.Error
}
