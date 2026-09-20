package books

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/cities"
	"github.com/kaellybot/kaelly-discord/repositories/jobs"
	"github.com/kaellybot/kaelly-discord/repositories/orders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeJobRepo struct {
	jobs.Repository
	jobs  []entities.Job
	err   error
	calls int
}

func (repo *fakeJobRepo) GetJobs() ([]entities.Job, error) {
	repo.calls++
	return repo.jobs, repo.err
}

type fakeCityRepo struct {
	cities.Repository
	cities []entities.City
	err    error
	calls  int
}

func (repo *fakeCityRepo) GetCities() ([]entities.City, error) {
	repo.calls++
	return repo.cities, repo.err
}

type fakeOrderRepo struct {
	orders.Repository
	orders []entities.Order
	err    error
	calls  int
}

func (repo *fakeOrderRepo) GetOrders() ([]entities.Order, error) {
	repo.calls++
	return repo.orders, repo.err
}

func job(id, fr, en string) entities.Job {
	return entities.Job{ID: id, Labels: []entities.JobLabel{
		{JobID: id, Locale: amqp.Language_FR, Label: fr},
		{JobID: id, Locale: amqp.Language_EN, Label: en},
	}}
}

func city(id, fr, en string) entities.City {
	return entities.City{ID: id, Labels: []entities.CityLabel{
		{CityID: id, Locale: amqp.Language_FR, Label: fr},
		{CityID: id, Locale: amqp.Language_EN, Label: en},
	}}
}

func order(id, fr, en string) entities.Order {
	return entities.Order{ID: id, Labels: []entities.OrderLabel{
		{OrderID: id, Locale: amqp.Language_FR, Label: fr},
		{OrderID: id, Locale: amqp.Language_EN, Label: en},
	}}
}

type labelled interface{ GetID() string }

func ids[T labelled](items []T) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.GetID())
	}
	return out
}

func newService(t *testing.T) *Impl {
	t.Helper()
	service, err := New(
		&fakeJobRepo{jobs: []entities.Job{
			job("lumberjack", "Bûcheron", "Lumberjack"),
			job("jeweller", "Bijoutier", "Jeweller"),
			job("shoemaker", "Cordonnier", "Shoemaker"),
			job("shoemagus", "Cordomage", "Shoemagus"),
		}},
		&fakeCityRepo{cities: []entities.City{
			city("bonta", "Bonta", "Bonta"),
			city("brakmar", "Brâkmar", "Brakmar"),
		}},
		&fakeOrderRepo{orders: []entities.Order{
			order("heart", "Cœur", "Heart"),
			order("spirit", "Esprit", "Spirit"),
			order("spirit-2", "Esprit Bis", "Spirit Bis"),
		}},
	)
	require.NoError(t, err)
	return service
}

func TestNew_StopsAtFirstRepositoryError(t *testing.T) {
	boom := errors.New("db down")

	t.Run("jobs", func(t *testing.T) {
		jobRepo, cityRepo, orderRepo := &fakeJobRepo{err: boom}, &fakeCityRepo{}, &fakeOrderRepo{}
		service, err := New(jobRepo, cityRepo, orderRepo)
		require.ErrorIs(t, err, boom)
		assert.Nil(t, service)
		assert.Zero(t, cityRepo.calls, "cities are not loaded once jobs failed")
		assert.Zero(t, orderRepo.calls)
	})

	t.Run("cities", func(t *testing.T) {
		jobRepo, cityRepo, orderRepo := &fakeJobRepo{}, &fakeCityRepo{err: boom}, &fakeOrderRepo{}
		service, err := New(jobRepo, cityRepo, orderRepo)
		require.ErrorIs(t, err, boom)
		assert.Nil(t, service)
		assert.Equal(t, 1, jobRepo.calls)
		assert.Zero(t, orderRepo.calls, "orders are not loaded once cities failed")
	})

	t.Run("orders", func(t *testing.T) {
		jobRepo, cityRepo, orderRepo := &fakeJobRepo{}, &fakeCityRepo{}, &fakeOrderRepo{err: boom}
		service, err := New(jobRepo, cityRepo, orderRepo)
		require.ErrorIs(t, err, boom)
		assert.Nil(t, service)
		assert.Equal(t, 1, jobRepo.calls)
		assert.Equal(t, 1, cityRepo.calls)
	})
}

func TestGetters_ExposeLoadedEntitiesByIDAndInOrder(t *testing.T) {
	service := newService(t)

	assert.Equal(t, []string{"lumberjack", "jeweller", "shoemaker", "shoemagus"}, ids(service.GetJobs()))
	assert.Equal(t, []string{"bonta", "brakmar"}, ids(service.GetCities()))
	assert.Equal(t, []string{"heart", "spirit", "spirit-2"}, ids(service.GetOrders()))

	gotJob, found := service.GetJob("jeweller")
	assert.True(t, found)
	assert.Equal(t, "jeweller", gotJob.ID)
	_, found = service.GetJob("bonta")
	assert.False(t, found, "each getter only looks into its own entity kind")

	gotCity, found := service.GetCity("brakmar")
	assert.True(t, found)
	assert.Equal(t, "brakmar", gotCity.ID)
	_, found = service.GetCity("heart")
	assert.False(t, found)

	gotOrder, found := service.GetOrder("spirit")
	assert.True(t, found)
	assert.Equal(t, "spirit", gotOrder.ID)
	_, found = service.GetOrder("lumberjack")
	assert.False(t, found)
}

func TestFindJobs(t *testing.T) {
	service := newService(t)

	assert.Equal(t, []string{"lumberjack"}, ids(service.FindJobs("BUCHE", discordgo.French, 10)),
		"accents and case are ignored")
	assert.Equal(t, []string{"shoemaker", "shoemagus"}, ids(service.FindJobs("cordo", discordgo.French, 10)))
	assert.Equal(t, []string{"shoemaker"}, ids(service.FindJobs("cordo", discordgo.French, 1)),
		"partials are truncated to the limit")
	assert.Empty(t, service.FindJobs("cordo", discordgo.French, 0))
	assert.Empty(t, service.FindJobs("cordo", discordgo.French, -1), "a negative limit yields nothing")
	assert.Equal(t, []string{"shoemagus"}, ids(service.FindJobs("Shoemagus", discordgo.EnglishGB, 10)),
		"exact match wins over the earlier partial candidates")
	assert.Empty(t, service.FindJobs("bucheron", discordgo.EnglishUS, 10), "French label not searched in English")
	assert.Equal(t, []string{"lumberjack"}, ids(service.FindJobs("lumber", discordgo.Japanese, 10)),
		"unsupported locale searches English labels")
	assert.NotNil(t, service.FindJobs("zzz", discordgo.French, 10))
}

func TestFindCities(t *testing.T) {
	service := newService(t)

	assert.Equal(t, []string{"brakmar"}, ids(service.FindCities("brakmar", discordgo.French, 0)),
		"accent-insensitive exact match bypasses the limit")
	assert.Equal(t, []string{"bonta", "brakmar"}, ids(service.FindCities("B", discordgo.French, 10)))
	assert.Equal(t, []string{"bonta"}, ids(service.FindCities("b", discordgo.French, 1)))
	assert.Empty(t, service.FindCities("b", discordgo.French, 0))
	assert.Empty(t, service.FindCities("b", discordgo.French, -1), "a negative limit yields nothing")
	none := service.FindCities("amakna", discordgo.French, 10)
	assert.NotNil(t, none)
	assert.Empty(t, none)
}

func TestFindOrders(t *testing.T) {
	service := newService(t)

	assert.Equal(t, []string{"spirit"}, ids(service.FindOrders("ESPRIT", discordgo.French, 10)),
		"exact match wins over the partial 'Esprit Bis' that follows it")
	assert.Equal(t, []string{"spirit", "spirit-2"}, ids(service.FindOrders("spir", discordgo.EnglishGB, 10)))
	assert.Equal(t, []string{"heart"}, ids(service.FindOrders("cœ", discordgo.French, 10)))
	assert.Equal(t, []string{"heart"}, ids(service.FindOrders("coeur", discordgo.French, 10)),
		"the œ ligature is expanded, so typing 'coeur' matches 'Cœur'")
	assert.Equal(t, []string{"heart"}, ids(service.FindOrders("COEUR", discordgo.French, 10)))
	assert.Empty(t, service.FindOrders("spir", discordgo.EnglishGB, -1), "a negative limit yields nothing")
	assert.NotNil(t, service.FindOrders("zzz", discordgo.French, 10))
}
