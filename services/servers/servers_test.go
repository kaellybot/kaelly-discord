package servers

import (
	"errors"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/servers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeServerRepo struct {
	repository.Repository
	servers []entities.Server
	err     error
}

func (repo *fakeServerRepo) GetServers() ([]entities.Server, error) {
	return repo.servers, repo.err
}

// server builds a server labelled in French and English.
func server(id, fr, en string) entities.Server {
	return entities.Server{ID: id, Labels: []entities.ServerLabel{
		{Locale: amqp.Language_FR, ServerID: id, Label: fr},
		{Locale: amqp.Language_EN, ServerID: id, Label: en},
	}}
}

func newService(t *testing.T, servers ...entities.Server) *Impl {
	t.Helper()
	service, err := New(&fakeServerRepo{servers: servers})
	require.NoError(t, err)
	return service
}

func ids(servers []entities.Server) []string {
	out := make([]string, 0, len(servers))
	for _, s := range servers {
		out = append(out, s.ID)
	}
	return out
}

func TestNew_PropagatesRepositoryError(t *testing.T) {
	boom := errors.New("db down")
	service, err := New(&fakeServerRepo{err: boom})
	require.ErrorIs(t, err, boom)
	assert.Nil(t, service)
}

func TestGetServers_KeepsRepositoryOrder(t *testing.T) {
	service := newService(t, server("b", "B", "B"), server("a", "A", "A"))
	assert.Equal(t, []string{"b", "a"}, ids(service.GetServers()))
}

func TestGetServer_LooksUpByID(t *testing.T) {
	service := newService(t, server("dakal", "Dakal", "Dakal"))

	got, found := service.GetServer("dakal")
	require.True(t, found)
	assert.Equal(t, "dakal", got.ID)

	got, found = service.GetServer("Dakal")
	assert.False(t, found, "ids are case-sensitive, unlike FindServers")
	assert.Zero(t, got)
}

func TestFindServers_IgnoresCaseAndAccentsOnBothSides(t *testing.T) {
	service := newService(t,
		server("heroic", "Héroïque", "Heroic"),
		server("other", "Autre", "Other"),
	)

	for _, query := range []string{"heroiq", "HÉROÏ", "éroi", "HeRoÏqUe"} {
		t.Run(query, func(t *testing.T) {
			assert.Equal(t, []string{"heroic"}, ids(service.FindServers(query, discordgo.French, 10)))
		})
	}
}

func TestFindServers_ExactMatchWinsOverPartials(t *testing.T) {
	service := newService(t,
		server("dakal-2", "Dakal Deux", "Dakal Two"),
		server("dakal", "Dakal", "Dakal"),
		server("dakal-3", "Dakal Trois", "Dakal Three"),
	)

	assert.Equal(t, []string{"dakal"}, ids(service.FindServers("DÂKAL", discordgo.French, 10)),
		"a normalized exact match discards partials found before and after it")
}

func TestFindServers_ExactMatchIgnoresLimit(t *testing.T) {
	service := newService(t, server("dakal", "Dakal", "Dakal"))
	assert.Equal(t, []string{"dakal"}, ids(service.FindServers("dakal", discordgo.French, 0)),
		"the exact-match short-circuit returns before the limit is applied")
}

func TestFindServers_PartialMatchesAreTruncatedToLimitInRepositoryOrder(t *testing.T) {
	service := newService(t,
		server("s1", "Serveur Un", "Server One"),
		server("s2", "Serveur Deux", "Server Two"),
		server("x", "Autre", "Other"),
		server("s3", "Serveur Trois", "Server Three"),
	)

	assert.Equal(t, []string{"s1", "s2"}, ids(service.FindServers("serveur", discordgo.French, 2)))
	assert.Equal(t, []string{"s1", "s2", "s3"}, ids(service.FindServers("serveur", discordgo.French, 3)))
	assert.Equal(t, []string{"s1", "s2", "s3"}, ids(service.FindServers("serveur", discordgo.French, 25)))
	assert.Empty(t, service.FindServers("serveur", discordgo.French, 0))
}

func TestFindServers_EmptyQueryListsEverythingUpToLimit(t *testing.T) {
	service := newService(t, server("a", "A1", "A1"), server("b", "B1", "B1"), server("c", "C1", "C1"))
	assert.Equal(t, []string{"a", "b"}, ids(service.FindServers("", discordgo.French, 2)))
}

func TestFindServers_NoMatch_ReturnsEmptyNonNilSlice(t *testing.T) {
	service := newService(t, server("a", "Alpha", "Alpha"))
	found := service.FindServers("zeta", discordgo.French, 10)
	assert.NotNil(t, found, "callers range over it / serialize it")
	assert.Empty(t, found)
}

func TestFindServers_SearchesInTheCallerLocaleOnly(t *testing.T) {
	service := newService(t, server("rushu", "Épée", "Sword"))

	assert.Equal(t, []string{"rushu"}, ids(service.FindServers("epee", discordgo.French, 10)))
	assert.Empty(t, service.FindServers("sword", discordgo.French, 10), "English label is not searched for French")
	assert.Equal(t, []string{"rushu"}, ids(service.FindServers("sword", discordgo.EnglishUS, 10)))
	assert.Empty(t, service.FindServers("epee", discordgo.EnglishGB, 10))
}

func TestFindServers_FallsBackToDefaultLocaleThenID(t *testing.T) {
	onlyEnglish := entities.Server{ID: "en-only", Labels: []entities.ServerLabel{
		{Locale: amqp.Language_EN, Label: "Draconiros"},
	}}
	unlabelled := entities.Server{ID: "tylezia"}
	service := newService(t, onlyEnglish, unlabelled)

	assert.Equal(t, []string{"en-only"}, ids(service.FindServers("draco", discordgo.German, 10)),
		"missing German label falls back to the default (English) label")
	assert.Equal(t, []string{"en-only"}, ids(service.FindServers("draco", discordgo.Japanese, 10)),
		"unsupported locale falls back to the default (English) label")
	assert.Equal(t, []string{"tylezia"}, ids(service.FindServers("TYLEZ", discordgo.French, 10)),
		"no label at all: the id is searched")
}

func TestFindServers_ConcurrentSearchesReturnTheSameMatch(t *testing.T) {
	// Autocompletes arrive on concurrent goroutines: the search used to share
	// one stateful transformer, so parallel calls corrupted each other's text.
	service := newService(t,
		server("henual", "Hénual", "Henual"),
		server("crapaud", "Crâpaud", "Crapaud"),
		server("imagiro", "Ìmagiro", "Imagiro"))

	const goroutines = 32
	var waitGroup sync.WaitGroup
	results := make([][]string, goroutines)

	for index := range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			results[index] = ids(service.FindServers("crapaud", discordgo.French, 10))
		}()
	}
	waitGroup.Wait()

	for _, got := range results {
		assert.Equal(t, []string{"crapaud"}, got)
	}
}
