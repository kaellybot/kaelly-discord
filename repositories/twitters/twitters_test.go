package twitters

import (
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTwitterAccounts_ReturnsAccountsOfEveryGame(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t, &entities.TwitterAccount{})
	accounts := []entities.TwitterAccount{
		{ID: "dofusfr", Name: "DOFUSfr", NewsChannelID: "c1", Game: amqp.Game_DOFUS_GAME},
		{ID: "dofustouch", Name: "DOFUSTouch", NewsChannelID: "c2", Game: amqp.Game_DOFUS_TOUCH},
		{ID: "ankama", Name: "Ankama", NewsChannelID: "c3", Game: amqp.Game_ANY_GAME},
	}
	require.NoError(t, conn.DB.Create(&accounts).Error)

	got, err := New(conn).GetTwitterAccounts()

	require.NoError(t, err)
	// Unlike the other game-scoped repositories, no game filter is applied.
	assert.ElementsMatch(t, accounts, got)
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
