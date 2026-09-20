package feeds

import (
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newConn(t *testing.T) *testutils.SQLiteConnection {
	t.Helper()
	return testutils.NewSQLiteConnection(t,
		&entities.FeedType{}, &entities.FeedTypeLabel{}, &entities.FeedSource{})
}

func TestGetFeedTypes_PreloadsAllLabelsButOnlyCurrentGameSources(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&[]entities.FeedType{{ID: "news"}, {ID: "devblog"}}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.FeedTypeLabel{
		{FeedTypeID: "news", Locale: amqp.Language_FR, Label: "Actualités"},
		{FeedTypeID: "news", Locale: amqp.Language_EN, Label: "News"},
		{FeedTypeID: "devblog", Locale: amqp.Language_FR, Label: "Devblog"},
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.FeedSource{
		{FeedTypeID: "news", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, NewsChannelID: "news-fr"},
		{FeedTypeID: "news", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_EN, NewsChannelID: "news-en"},
		{FeedTypeID: "news", Game: amqp.Game_DOFUS_TOUCH, Locale: amqp.Language_FR, NewsChannelID: "news-touch"},
		{FeedTypeID: "devblog", Game: amqp.Game_DOFUS_RETRO, Locale: amqp.Language_FR, NewsChannelID: "dev-retro"},
	}).Error)

	got, err := New(conn).GetFeedTypes()

	require.NoError(t, err)
	require.Len(t, got, 2)
	byID := map[string]entities.FeedType{}
	for _, feedType := range got {
		byID[feedType.ID] = feedType
	}

	news := byID["news"]
	assert.ElementsMatch(t, []entities.FeedTypeLabel{
		{FeedTypeID: "news", Locale: amqp.Language_FR, Label: "Actualités"},
		{FeedTypeID: "news", Locale: amqp.Language_EN, Label: "News"},
	}, news.Labels)
	assert.ElementsMatch(t, []entities.FeedSource{
		{FeedTypeID: "news", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, NewsChannelID: "news-fr"},
		{FeedTypeID: "news", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_EN, NewsChannelID: "news-en"},
	}, news.Sources)

	// Feed types themselves are not filtered by game: a type with no source
	// for the current game is still returned, with no sources.
	devblog := byID["devblog"]
	assert.Equal(t, []entities.FeedTypeLabel{
		{FeedTypeID: "devblog", Locale: amqp.Language_FR, Label: "Devblog"},
	}, devblog.Labels)
	assert.Empty(t, devblog.Sources)
}

func TestGetFeedTypes_EmptyTable(t *testing.T) {
	got, err := New(newConn(t)).GetFeedTypes()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetFeedTypes_PropagatesErrors(t *testing.T) {
	for _, table := range []string{"feed_types", "feed_type_labels", "feed_sources"} {
		t.Run(table+" missing", func(t *testing.T) {
			conn := newConn(t)
			require.NoError(t, conn.DB.Create(&entities.FeedType{ID: "news"}).Error)
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			_, err := New(conn).GetFeedTypes()

			require.Error(t, err)
		})
	}
}
