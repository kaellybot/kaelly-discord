package jobs

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type JobRepository interface {
	GetJobs() ([]entities.Job, error)
}

type JobRepositoryImpl struct {
	db databases.MySQLConnection
}

func New(db databases.MySQLConnection) *JobRepositoryImpl {
	return &JobRepositoryImpl{db: db}
}

func (repo *JobRepositoryImpl) GetJobs() ([]entities.Job, error) {
	var jobs []entities.Job
	response := repo.db.GetDB().Model(&entities.Job{}).Preload("Labels").Find(&jobs)
	return jobs, response.Error
}
