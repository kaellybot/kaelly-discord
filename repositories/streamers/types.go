package streamers

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetStreamers() ([]entities.Streamer, error)
}

type Impl struct {
	db databases.MySQLConnection
}
