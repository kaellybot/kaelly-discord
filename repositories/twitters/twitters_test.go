package twitters

import (
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Each game has its own accounts and its own news channel to follow, so this bot
// must only ever see its own: following another game's channel would make it
// mirror that bot's news.
func TestGetTwitterAccounts_OnlyReturnsThisGamesAccounts(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t, &entities.TwitterAccount{})
	ours := constants.GetGame().AMQPGame
	theirs := amqp.Game_DOFUS_TOUCH
	if ours == theirs {
		theirs = amqp.Game_DOFUS_GAME
	}

	require.NoError(t, conn.DB.Create(&[]entities.TwitterAccount{
		{ID: "dofusfr", Game: ours, Name: "DOFUSfr", NewsChannelID: "c1"},
		// Same account ID under the other game, and an account of the other game.
		{ID: "dofusfr", Game: theirs, Name: "DOFUSfr", NewsChannelID: "c2"},
		{ID: "dofustouch", Game: theirs, Name: "DOFUSTouch", NewsChannelID: "c3"},
		{ID: "unset", Game: amqp.Game_ANY_GAME, Name: "Unset", NewsChannelID: "c4"},
	}).Error)

	got, err := New(conn).GetTwitterAccounts()

	require.NoError(t, err)
	assert.ElementsMatch(t, []entities.TwitterAccount{
		{ID: "dofusfr", Game: ours, Name: "DOFUSfr", NewsChannelID: "c1"},
	}, got)
}

func TestGetTwitterAccounts_EmptyTable(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t, &entities.TwitterAccount{})

	got, err := New(conn).GetTwitterAccounts()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetTwitterAccounts_PropagatesDatabaseError(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t)

	got, err := New(conn).GetTwitterAccounts()

	require.Error(t, err)
	assert.Empty(t, got)
}
