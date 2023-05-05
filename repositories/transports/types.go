package transports

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetTransportTypes() ([]entities.TransportType, error)
}

type Impl struct {
	db databases.MySQLConnection
}
