package transports

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetTransportTypes() ([]entities.TransportType, error) {
	var transportTypes []entities.TransportType
	response := repo.db.GetDB().Model(&entities.TransportType{}).Preload("Labels").Find(&transportTypes)
	return transportTypes, response.Error
}
