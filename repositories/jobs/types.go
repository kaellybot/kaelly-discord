package jobs

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetJobs() ([]entities.Job, error)
}

type Impl struct {
	db databases.MySQLConnection
}
