package almanaxes

import (
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAlmanaxNews_OnlyReturnsCurrentGameChannels(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t, &entities.AlmanaxNews{})
	require.NoError(t, conn.DB.Create(&[]entities.AlmanaxNews{
		{Locale: amqp.Language_FR, Game: amqp.Game_DOFUS_GAME, NewsChannelID: "fr-dofus"},
		{Locale: amqp.Language_EN, Game: amqp.Game_DOFUS_GAME, NewsChannelID: "en-dofus"},
		{Locale: amqp.Language_FR, Game: amqp.Game_DOFUS_TOUCH, NewsChannelID: "fr-touch"},
		{Locale: amqp.Language_FR, Game: amqp.Game_ANY_GAME, NewsChannelID: "fr-any"},
	}).Error)

	got, err := New(conn).GetAlmanaxNews()

	require.NoError(t, err)
	assert.ElementsMatch(t, []entities.AlmanaxNews{
		{Locale: amqp.Language_FR, Game: amqp.Game_DOFUS_GAME, NewsChannelID: "fr-dofus"},
		{Locale: amqp.Language_EN, Game: amqp.Game_DOFUS_GAME, NewsChannelID: "en-dofus"},
	}, got)
}

func TestGetAlmanaxNews_EmptyWhenNoRowForCurrentGame(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t, &entities.AlmanaxNews{})
	require.NoError(t, conn.DB.Create(&entities.AlmanaxNews{
		Locale: amqp.Language_FR, Game: amqp.Game_DOFUS_RETRO, NewsChannelID: "fr-retro",
	}).Error)

	got, err := New(conn).GetAlmanaxNews()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetAlmanaxNews_PropagatesDatabaseError(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t) // table never created

	got, err := New(conn).GetAlmanaxNews()

	require.Error(t, err)
	assert.Empty(t, got)
}
