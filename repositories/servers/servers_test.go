package servers

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
	return testutils.NewSQLiteConnection(t, &entities.Server{}, &entities.ServerLabel{})
}

func TestGetServers_FiltersByGameAndPreloadsLabels(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&[]entities.Server{
		{ID: "draconiros", Icon: "d.webp", Image: "d.png", Game: amqp.Game_DOFUS_GAME},
		{ID: "imagiro", Icon: "i.webp", Image: "i.png", Game: amqp.Game_DOFUS_GAME},
		{ID: "tiliwan", Game: amqp.Game_DOFUS_TOUCH},
		{ID: "boune", Game: amqp.Game_DOFUS_RETRO},
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.ServerLabel{
		{ServerID: "draconiros", Locale: amqp.Language_FR, Label: "Draconiros"},
		{ServerID: "draconiros", Locale: amqp.Language_ES, Label: "Draconiros ES"},
		{ServerID: "tiliwan", Locale: amqp.Language_FR, Label: "Tiliwan"},
	}).Error)

	got, err := New(conn).GetServers()

	require.NoError(t, err)
	require.Len(t, got, 2)
	byID := map[string]entities.Server{}
	for _, server := range got {
		assert.Equal(t, amqp.Game_DOFUS_GAME, server.Game)
		byID[server.ID] = server
	}
	assert.Equal(t, "d.webp", byID["draconiros"].Icon)
	assert.Equal(t, "d.png", byID["draconiros"].Image)
	assert.ElementsMatch(t, []entities.ServerLabel{
		{ServerID: "draconiros", Locale: amqp.Language_FR, Label: "Draconiros"},
		{ServerID: "draconiros", Locale: amqp.Language_ES, Label: "Draconiros ES"},
	}, byID["draconiros"].Labels)
	assert.Empty(t, byID["imagiro"].Labels)
}

func TestGetServers_EmptyWhenNoServerForCurrentGame(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&entities.Server{ID: "tiliwan", Game: amqp.Game_DOFUS_TOUCH}).Error)

	got, err := New(conn).GetServers()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetServers_PropagatesErrors(t *testing.T) {
	for _, table := range []string{"servers", "server_labels"} {
		t.Run(table+" missing", func(t *testing.T) {
			conn := newConn(t)
			require.NoError(t, conn.DB.Create(&entities.Server{ID: "imagiro", Game: amqp.Game_DOFUS_GAME}).Error)
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			_, err := New(conn).GetServers()

			require.Error(t, err)
		})
	}
}
