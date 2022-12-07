package servers

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type ServerRepository interface {
	GetServers() ([]entities.Server, error)
}

type ServerRepositoryImpl struct {
	db databases.MySQLConnection
}

func New(db databases.MySQLConnection) *ServerRepositoryImpl {
	return &ServerRepositoryImpl{db: db}
}

func (repo *ServerRepositoryImpl) GetServers() ([]entities.Server, error) {
	var servers []entities.Server
	response := repo.db.GetDB().Find(&servers)
	return servers, response.Error
}
