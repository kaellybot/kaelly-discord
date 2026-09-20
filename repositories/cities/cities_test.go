package cities

import (
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newConn(t *testing.T) *testutils.SQLiteConnection {
	t.Helper()
	return testutils.NewSQLiteConnection(t, &entities.City{}, &entities.CityLabel{})
}

func TestGetCities_FiltersByGameAndPreloadsLabelsOfSameGame(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&[]entities.City{
		{ID: "bonta", Icon: "bonta.webp", Type: constants.CityTypeLight, Color: 1, Game: amqp.Game_DOFUS_GAME},
		{ID: "brakmar", Icon: "brak.webp", Type: constants.CityTypeDark, Color: 2, Game: amqp.Game_DOFUS_GAME},
		// Same ID, other game: must be excluded, and its labels must not leak.
		{ID: "bonta", Icon: "touch.webp", Type: constants.CityTypeLight, Color: 3, Game: amqp.Game_DOFUS_TOUCH},
		{ID: "retro-only", Game: amqp.Game_DOFUS_RETRO},
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.CityLabel{
		{CityID: "bonta", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, Label: "Bonta"},
		{CityID: "bonta", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_EN, Label: "Bonta EN"},
		{CityID: "bonta", Game: amqp.Game_DOFUS_TOUCH, Locale: amqp.Language_FR, Label: "Bonta Touch"},
		{CityID: "brakmar", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, Label: "Brâkmar"},
	}).Error)

	got, err := New(conn).GetCities()

	require.NoError(t, err)
	require.Len(t, got, 2)
	byID := map[string]entities.City{}
	for _, city := range got {
		assert.Equal(t, amqp.Game_DOFUS_GAME, city.Game)
		byID[city.ID] = city
	}
	bonta := byID["bonta"]
	assert.Equal(t, "bonta.webp", bonta.Icon)
	assert.Equal(t, constants.CityTypeLight, bonta.Type)
	assert.Equal(t, 1, bonta.Color)
	assert.ElementsMatch(t, []entities.CityLabel{
		{CityID: "bonta", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, Label: "Bonta"},
		{CityID: "bonta", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_EN, Label: "Bonta EN"},
	}, bonta.Labels)
	brakmar := byID["brakmar"]
	assert.Equal(t, constants.CityTypeDark, brakmar.Type)
	assert.Equal(t, []entities.CityLabel{
		{CityID: "brakmar", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, Label: "Brâkmar"},
	}, brakmar.Labels)
}

func TestGetCities_EmptyWhenNoCityForCurrentGame(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&entities.City{ID: "bonta", Game: amqp.Game_DOFUS_TOUCH}).Error)

	got, err := New(conn).GetCities()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetCities_PropagatesErrors(t *testing.T) {
	for _, table := range []string{"cities", "city_labels"} {
		t.Run(table+" missing", func(t *testing.T) {
			conn := newConn(t)
			require.NoError(t, conn.DB.Create(&entities.City{ID: "bonta", Game: amqp.Game_DOFUS_GAME}).Error)
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			_, err := New(conn).GetCities()

			require.Error(t, err)
		})
	}
}
