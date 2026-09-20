package weapons

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
	return testutils.NewSQLiteConnection(t, &entities.WeaponAreaEffect{}, &entities.WeaponAreaEffectLabel{})
}

func TestGetWeaponAreaEffects_PreloadsLabelsOfEachEffect(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&[]entities.WeaponAreaEffect{
		{ID: "cross", Type: constants.Area, Order: 2},
		{ID: "line", Type: constants.Casting, Order: 1},
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.WeaponAreaEffectLabel{
		{WeaponAreaEffectID: "cross", Locale: amqp.Language_FR, Label: "Croix"},
		{WeaponAreaEffectID: "cross", Locale: amqp.Language_EN, Label: "Cross"},
		{WeaponAreaEffectID: "line", Locale: amqp.Language_FR, Label: "Ligne"},
		{WeaponAreaEffectID: "unknown", Locale: amqp.Language_FR, Label: "Inconnu"},
	}).Error)

	got, err := New(conn).GetWeaponAreaEffects()

	require.NoError(t, err)
	require.Len(t, got, 2)
	byID := map[string]entities.WeaponAreaEffect{}
	for _, effect := range got {
		byID[effect.ID] = effect
	}
	assert.Equal(t, constants.Area, byID["cross"].Type)
	assert.Equal(t, 2, byID["cross"].Order)
	assert.ElementsMatch(t, []entities.WeaponAreaEffectLabel{
		{WeaponAreaEffectID: "cross", Locale: amqp.Language_FR, Label: "Croix"},
		{WeaponAreaEffectID: "cross", Locale: amqp.Language_EN, Label: "Cross"},
	}, byID["cross"].Labels)
	assert.Equal(t, constants.Casting, byID["line"].Type)
	assert.Equal(t, []entities.WeaponAreaEffectLabel{
		{WeaponAreaEffectID: "line", Locale: amqp.Language_FR, Label: "Ligne"},
	}, byID["line"].Labels)
}

func TestGetWeaponAreaEffects_EmptyTable(t *testing.T) {
	got, err := New(newConn(t)).GetWeaponAreaEffects()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetWeaponAreaEffects_PropagatesErrors(t *testing.T) {
	for _, table := range []string{"weapon_area_effects", "weapon_area_effect_labels"} {
		t.Run(table+" missing", func(t *testing.T) {
			conn := newConn(t)
			require.NoError(t, conn.DB.Create(&entities.WeaponAreaEffect{ID: "line", Type: constants.Casting}).Error)
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			_, err := New(conn).GetWeaponAreaEffects()

			require.Error(t, err)
		})
	}
}
