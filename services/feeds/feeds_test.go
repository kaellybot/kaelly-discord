package feeds

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/feeds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFeedRepo struct {
	repository.Repository
	feedTypes []entities.FeedType
	err       error
}

func (repo *fakeFeedRepo) GetFeedTypes() ([]entities.FeedType, error) {
	return repo.feedTypes, repo.err
}

func feedType(id, fr, en string) entities.FeedType {
	return entities.FeedType{ID: id, Labels: []entities.FeedTypeLabel{
		{Locale: amqp.Language_FR, FeedTypeID: id, Label: fr},
		{Locale: amqp.Language_EN, FeedTypeID: id, Label: en},
	}}
}

func newService(t *testing.T, feedTypes ...entities.FeedType) *Impl {
	t.Helper()
	service, err := New(&fakeFeedRepo{feedTypes: feedTypes})
	require.NoError(t, err)
	return service
}

func ids(feedTypes []entities.FeedType) []string {
	out := make([]string, 0, len(feedTypes))
	for _, f := range feedTypes {
		out = append(out, f.ID)
	}
	return out
}

func TestNew_PropagatesRepositoryError(t *testing.T) {
	boom := errors.New("db down")
	service, err := New(&fakeFeedRepo{err: boom})
	require.ErrorIs(t, err, boom)
	assert.Nil(t, service)
}

func TestGetFeedTypes_KeepsRepositoryOrder(t *testing.T) {
	service := newService(t, feedType("news", "Actualités", "News"), feedType("devblog", "Devblog", "Devblog"))
	assert.Equal(t, []string{"news", "devblog"}, ids(service.GetFeedTypes()))
}

func TestGetFeedType_ReturnsCopyOrNil(t *testing.T) {
	service := newService(t, feedType("news", "Actualités", "News"))

	got := service.GetFeedType("news")
	require.NotNil(t, got)
	assert.Equal(t, "news", got.ID)

	got.ID = "mutated"
	assert.NotNil(t, service.GetFeedType("news"), "the cached feed type is not mutated through the pointer")

	assert.Nil(t, service.GetFeedType("NEWS"), "ids are case-sensitive")
	assert.Nil(t, service.GetFeedType("unknown"))
}

func TestFindFeedTypes_IgnoresCaseAndAccents(t *testing.T) {
	service := newService(t,
		feedType("news", "Actualités", "News"),
		feedType("patch", "Mises à jour", "Updates"),
	)
	assert.Equal(t, []string{"news"}, ids(service.FindFeedTypes("ACTUALITE", discordgo.French, 10)))
	assert.Equal(t, []string{"patch"}, ids(service.FindFeedTypes("MISES A", discordgo.French, 10)))
}

func TestFindFeedTypes_ExactMatchWinsOverPartials(t *testing.T) {
	service := newService(t,
		feedType("news-pc", "Actualités PC", "News PC"),
		feedType("news", "Actualités", "News"),
	)
	assert.Equal(t, []string{"news"}, ids(service.FindFeedTypes("news", discordgo.EnglishGB, 10)))
	assert.Equal(t, []string{"news"}, ids(service.FindFeedTypes("news", discordgo.EnglishGB, 0)),
		"exact match bypasses the limit")
}

func TestFindFeedTypes_LimitAndLocale(t *testing.T) {
	service := newService(t,
		feedType("a", "Flux A", "Feed A"),
		feedType("b", "Flux B", "Feed B"),
		feedType("c", "Flux C", "Feed C"),
	)

	assert.Equal(t, []string{"a", "b"}, ids(service.FindFeedTypes("flux", discordgo.French, 2)))
	assert.Empty(t, service.FindFeedTypes("flux", discordgo.French, 0))
	assert.Empty(t, service.FindFeedTypes("flux", discordgo.French, -1), "a negative limit yields nothing")
	assert.Empty(t, service.FindFeedTypes("flux", discordgo.EnglishUS, 10), "French label not searched in English")
	assert.Equal(t, []string{"a", "b", "c"}, ids(service.FindFeedTypes("feed", discordgo.Japanese, 10)),
		"unsupported locale searches the default English labels")

	none := service.FindFeedTypes("zzz", discordgo.French, 10)
	assert.NotNil(t, none)
	assert.Empty(t, none)
}
