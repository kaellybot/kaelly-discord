package books

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/cities"
	"github.com/kaellybot/kaelly-discord/repositories/jobs"
	"github.com/kaellybot/kaelly-discord/repositories/orders"
	"golang.org/x/text/transform"
)

type Service interface {
	GetJob(id string) (entities.Job, bool)
	GetJobs() []entities.Job
	FindJobs(name string, locale discordgo.Locale) []entities.Job
	GetCity(id string) (entities.City, bool)
	GetCities() []entities.City
	FindCities(name string, locale discordgo.Locale) []entities.City
	GetOrder(id string) (entities.Order, bool)
	GetOrders() []entities.Order
	FindOrders(name string, locale discordgo.Locale) []entities.Order
}

type Impl struct {
	transformer     transform.Transformer
	jobsMap         map[string]entities.Job
	jobs            []entities.Job
	citiesMap       map[string]entities.City
	cities          []entities.City
	ordersMap       map[string]entities.Order
	orders          []entities.Order
	jobRepository   jobs.Repository
	cityRepository  cities.Repository
	orderRepository orders.Repository
}
