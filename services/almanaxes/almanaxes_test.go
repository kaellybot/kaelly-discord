package almanaxes

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/almanaxes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAlmanaxRepo struct {
	repository.Repository
	news  []entities.AlmanaxNews
	err   error
	calls int
}

func (repo *fakeAlmanaxRepo) GetAlmanaxNews() ([]entities.AlmanaxNews, error) {
	repo.calls++
	return repo.news, repo.err
}

func newService(t *testing.T, news ...entities.AlmanaxNews) *Impl {
	t.Helper()
	service, err := New(&fakeAlmanaxRepo{news: news})
	require.NoError(t, err)
	return service
}

func TestNew_PropagatesRepositoryError(t *testing.T) {
	boom := errors.New("db down")
	service, err := New(&fakeAlmanaxRepo{err: boom})
	require.ErrorIs(t, err, boom)
	assert.Nil(t, service)
}

func TestNew_LoadsNewsOnceAtStartup(t *testing.T) {
	repo := &fakeAlmanaxRepo{news: []entities.AlmanaxNews{
		{Locale: amqp.Language_FR, NewsChannelID: "fr-chan"},
	}}
	service, err := New(repo)
	require.NoError(t, err)

	service.GetAlmanaxNews(discordgo.French)
	service.GetAlmanaxNews(discordgo.French)
	assert.Equal(t, 1, repo.calls, "lookups are served from the in-memory cache")
}

func TestGetAlmanaxNews_MapsDiscordLocaleToLanguage(t *testing.T) {
	service := newService(t,
		entities.AlmanaxNews{Locale: amqp.Language_FR, NewsChannelID: "fr-chan"},
		entities.AlmanaxNews{Locale: amqp.Language_EN, NewsChannelID: "en-chan"},
		entities.AlmanaxNews{Locale: amqp.Language_PT, NewsChannelID: "pt-chan"},
	)

	cases := []struct {
		locale   discordgo.Locale
		expected string
	}{
		{discordgo.French, "fr-chan"},
		{discordgo.EnglishGB, "en-chan"},
		{discordgo.EnglishUS, "en-chan"},
		{discordgo.PortugueseBR, "pt-chan"},
	}
	for _, tc := range cases {
		t.Run(string(tc.locale), func(t *testing.T) {
			news := service.GetAlmanaxNews(tc.locale)
			require.NotNil(t, news)
			assert.Equal(t, tc.expected, news.NewsChannelID)
		})
	}
}

func TestGetAlmanaxNews_NoNewsForLocale_ReturnsNil(t *testing.T) {
	service := newService(t, entities.AlmanaxNews{Locale: amqp.Language_FR, NewsChannelID: "fr-chan"})
	assert.Nil(t, service.GetAlmanaxNews(discordgo.German), "supported locale without news")
}

func TestGetAlmanaxNews_UnsupportedLocale_DoesNotFallBackToDefault(t *testing.T) {
	service := newService(t, entities.AlmanaxNews{Locale: amqp.Language_EN, NewsChannelID: "en-chan"})
	assert.Nil(t, service.GetAlmanaxNews(discordgo.Japanese),
		"unsupported locales map to Language_ANY, which no news row carries")
}

func TestGetAlmanaxNews_ReturnsFirstMatchAsCopy(t *testing.T) {
	service := newService(t,
		entities.AlmanaxNews{Locale: amqp.Language_FR, Game: amqp.Game_DOFUS_GAME, NewsChannelID: "first"},
		entities.AlmanaxNews{Locale: amqp.Language_FR, Game: amqp.Game_DOFUS_TOUCH, NewsChannelID: "second"},
	)

	news := service.GetAlmanaxNews(discordgo.French)
	require.NotNil(t, news)
	assert.Equal(t, "first", news.NewsChannelID)

	news.NewsChannelID = "mutated"
	assert.Equal(t, "first", service.GetAlmanaxNews(discordgo.French).NewsChannelID,
		"callers cannot mutate the cached news through the returned pointer")
}
