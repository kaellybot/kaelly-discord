package books

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/cities"
	"github.com/kaellybot/kaelly-discord/repositories/jobs"
	"github.com/kaellybot/kaelly-discord/repositories/orders"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func New(jobRepository jobs.Repository, cityRepository cities.Repository,
	orderRepository orders.Repository) (*Impl, error) {
	jobs, errJob := jobRepository.GetJobs()
	if errJob != nil {
		return nil, errJob
	}

	log.Info().
		Int(constants.LogEntityCount, len(jobs)).
		Msgf("Jobs loaded")

	jobsMap := make(map[string]entities.Job)
	for _, job := range jobs {
		jobsMap[job.ID] = job
	}

	cities, errCity := cityRepository.GetCities()
	if errCity != nil {
		return nil, errCity
	}

	log.Info().
		Int(constants.LogEntityCount, len(cities)).
		Msgf("Cities loaded")

	citiesMap := make(map[string]entities.City)
	for _, city := range cities {
		citiesMap[city.ID] = city
	}

	orders, errOrder := orderRepository.GetOrders()
	if errOrder != nil {
		return nil, errOrder
	}

	log.Info().
		Int(constants.LogEntityCount, len(orders)).
		Msgf("Orders loaded")

	ordersMap := make(map[string]entities.Order)
	for _, order := range orders {
		ordersMap[order.ID] = order
	}

	return &Impl{
		transformer:     transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		jobsMap:         jobsMap,
		jobs:            jobs,
		citiesMap:       citiesMap,
		cities:          cities,
		ordersMap:       ordersMap,
		orders:          orders,
		jobRepository:   jobRepository,
		cityRepository:  cityRepository,
		orderRepository: orderRepository,
	}, nil
}

func (service *Impl) GetJobs() []entities.Job {
	return service.jobs
}

func (service *Impl) GetJob(id string) (entities.Job, bool) {
	job, found := service.jobsMap[id]
	return job, found
}

func (service *Impl) FindJobs(name string, locale discordgo.Locale) []entities.Job {
	jobsFound := make([]entities.Job, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize job name, returning empty job list")
		return jobsFound
	}

	for _, job := range service.jobs {
		currentCleanedName, _, errStr := transform.String(service.transformer,
			strings.ToLower(translators.GetEntityLabel(job, locale)))
		if errStr == nil && strings.HasPrefix(currentCleanedName, cleanedName) {
			jobsFound = append(jobsFound, job)
		}
	}

	return jobsFound
}

func (service *Impl) GetCity(id string) (entities.City, bool) {
	city, found := service.citiesMap[id]
	return city, found
}

func (service *Impl) GetCities() []entities.City {
	return service.cities
}

func (service *Impl) FindCities(name string, locale discordgo.Locale) []entities.City {
	citiesFound := make([]entities.City, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize city name, returning empty city list")
		return citiesFound
	}

	for _, city := range service.cities {
		currentCleanedName, _, errStr := transform.String(service.transformer,
			strings.ToLower(translators.GetEntityLabel(city, locale)))
		if errStr == nil && strings.HasPrefix(currentCleanedName, cleanedName) {
			citiesFound = append(citiesFound, city)
		}
	}

	return citiesFound
}

func (service *Impl) GetOrder(id string) (entities.Order, bool) {
	order, found := service.ordersMap[id]
	return order, found
}

func (service *Impl) GetOrders() []entities.Order {
	return service.orders
}

func (service *Impl) FindOrders(name string, locale discordgo.Locale) []entities.Order {
	ordersFound := make([]entities.Order, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize order name, returning empty order list")
		return ordersFound
	}

	for _, order := range service.orders {
		currentCleanedName, _, errStr := transform.String(service.transformer,
			strings.ToLower(translators.GetEntityLabel(order, locale)))
		if errStr == nil && strings.HasPrefix(currentCleanedName, cleanedName) {
			ordersFound = append(ordersFound, order)
		}
	}

	return ordersFound
}
