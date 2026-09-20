package characteristics

import (
	"testing"

	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newConn(t *testing.T) *testutils.SQLiteConnection {
	t.Helper()
	return testutils.NewSQLiteConnection(t, &entities.Characteristic{}, &entities.RegexCharacteristic{})
}

func TestGetCharacteristics_PreloadsRegexesOfEachCharacteristic(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&[]entities.Characteristic{
		{ID: "strength", SortOrder: 1},
		{ID: "wisdom", SortOrder: 2},
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.RegexCharacteristic{
		{CharacteristicID: "strength", Expression: "(?i)force"},
		{CharacteristicID: "strength", Expression: "(?i)str", RelativeCharacteristicID: "wisdom"},
		{CharacteristicID: "wisdom", Expression: "(?i)sagesse"},
		{CharacteristicID: "orphan", Expression: "(?i)nothing"},
	}).Error)

	got, err := New(conn).GetCharacteristics()

	require.NoError(t, err)
	require.Len(t, got, 2)
	byID := map[string]entities.Characteristic{}
	for _, c := range got {
		byID[c.ID] = c
	}
	assert.Equal(t, 1, byID["strength"].SortOrder)
	assert.ElementsMatch(t, []entities.RegexCharacteristic{
		{CharacteristicID: "strength", Expression: "(?i)force"},
		{CharacteristicID: "strength", Expression: "(?i)str", RelativeCharacteristicID: "wisdom"},
	}, byID["strength"].Regexes)
	assert.Equal(t, []entities.RegexCharacteristic{
		{CharacteristicID: "wisdom", Expression: "(?i)sagesse"},
	}, byID["wisdom"].Regexes)
}

func TestGetCharacteristics_CharacteristicWithoutRegexHasEmptyRegexes(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&entities.Characteristic{ID: "vitality", SortOrder: 1}).Error)

	got, err := New(conn).GetCharacteristics()

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "vitality", got[0].ID)
	assert.Empty(t, got[0].Regexes)
}

func TestGetCharacteristics_EmptyTable(t *testing.T) {
	got, err := New(newConn(t)).GetCharacteristics()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetCharacteristics_PropagatesErrors(t *testing.T) {
	cases := map[string]string{
		"main table missing":    "characteristics",
		"preload table missing": "regex_characteristics",
	}
	for name, table := range cases {
		t.Run(name, func(t *testing.T) {
			conn := newConn(t)
			require.NoError(t, conn.DB.Create(&entities.Characteristic{ID: "strength", SortOrder: 1}).Error)
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			_, err := New(conn).GetCharacteristics()

			require.Error(t, err)
		})
	}
}
