package books

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/jobs"
	"golang.org/x/text/transform"
)

type BookService interface {
	GetJob(id string) (entities.Job, bool)
	GetJobs() []entities.Job
	FindJobs(name string, locale discordgo.Locale) []entities.Job
}

type BookServiceImpl struct {
	transformer   transform.Transformer
	jobsMap       map[string]entities.Job
	jobs          []entities.Job
	jobRepository jobs.JobRepository
}
