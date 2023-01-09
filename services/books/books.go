package books

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/jobs"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func New(jobRepository jobs.JobRepository) (*BookServiceImpl, error) {
	jobs, err := jobRepository.GetJobs()
	if err != nil {
		return nil, err
	}

	jobsMap := make(map[string]entities.Job)
	for _, job := range jobs {
		jobsMap[job.Id] = job
	}

	return &BookServiceImpl{
		transformer:   transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		jobsMap:       jobsMap,
		jobs:          jobs,
		jobRepository: jobRepository,
	}, nil
}

func (service *BookServiceImpl) GetJobs() []entities.Job {
	return service.jobs
}

func (service *BookServiceImpl) GetJob(id string) (entities.Job, bool) {
	server, found := service.jobsMap[id]
	return server, found
}

func (service *BookServiceImpl) FindJobs(name string, locale discordgo.Locale) []entities.Job {
	jobsFound := make([]entities.Job, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize job name, returning empty job list")
		return jobsFound
	}

	for _, job := range service.jobs {
		currentCleanedName, _, err := transform.String(service.transformer, strings.ToLower(translators.GetEntityLabel(job, locale)))
		if err == nil && strings.HasPrefix(currentCleanedName, cleanedName) {
			jobsFound = append(jobsFound, job)
		}
	}

	return jobsFound
}
