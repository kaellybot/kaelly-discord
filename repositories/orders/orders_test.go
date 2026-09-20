package orders

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
	return testutils.NewSQLiteConnection(t, &entities.Order{}, &entities.OrderLabel{})
}

func TestGetOrders_FiltersByGameAndPreloadsLabelsOfSameGame(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&[]entities.Order{
		{ID: "heart", Game: amqp.Game_DOFUS_GAME},
		{ID: "spirit", Game: amqp.Game_DOFUS_GAME},
		{ID: "heart", Game: amqp.Game_DOFUS_RETRO},
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.OrderLabel{
		{OrderID: "heart", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, Label: "Coeur"},
		{OrderID: "heart", Game: amqp.Game_DOFUS_RETRO, Locale: amqp.Language_FR, Label: "Coeur Retro"},
		{OrderID: "spirit", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, Label: "Esprit"},
		{OrderID: "spirit", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_DE, Label: "Geist"},
	}).Error)

	got, err := New(conn).GetOrders()

	require.NoError(t, err)
	labels := map[string]map[amqp.Language]string{}
	for _, order := range got {
		assert.Equal(t, amqp.Game_DOFUS_GAME, order.Game)
		labels[order.ID] = order.GetLabels()
	}
	assert.Equal(t, map[string]map[amqp.Language]string{
		"heart":  {amqp.Language_FR: "Coeur"},
		"spirit": {amqp.Language_FR: "Esprit", amqp.Language_DE: "Geist"},
	}, labels)
}

func TestGetOrders_EmptyWhenNoOrderForCurrentGame(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&entities.Order{ID: "heart", Game: amqp.Game_DOFUS_TOUCH}).Error)

	got, err := New(conn).GetOrders()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetOrders_PropagatesErrors(t *testing.T) {
	for _, table := range []string{"orders", "order_labels"} {
		t.Run(table+" missing", func(t *testing.T) {
			conn := newConn(t)
			require.NoError(t, conn.DB.Create(&entities.Order{ID: "heart", Game: amqp.Game_DOFUS_GAME}).Error)
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			_, err := New(conn).GetOrders()

			require.Error(t, err)
		})
	}
}
