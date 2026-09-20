package emojis

import (
	"testing"

	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The production schema carries a snowflake_dev column that the entity does
// not map, so the table is created by hand.
func newConn(t *testing.T) *testutils.SQLiteConnection {
	t.Helper()
	conn := testutils.NewSQLiteConnection(t)
	require.NoError(t, conn.DB.Exec(`CREATE TABLE emojis (
		id TEXT, type TEXT, discord_name TEXT, name TEXT,
		snowflake TEXT, snowflake_dev TEXT,
		PRIMARY KEY (id, type))`).Error)
	require.NoError(t, conn.DB.Exec(`INSERT INTO emojis
		(id, type, discord_name, name, snowflake, snowflake_dev) VALUES
		('dofus', 'Item', 'dofus_emerald', 'Emerald', 'prod-1', 'dev-1'),
		('cape', 'Equipment', 'cape', 'Cape', 'prod-2', 'dev-2')`).Error)
	return conn
}

func setProduction(t *testing.T, value any) {
	t.Helper()
	previous := viper.Get(constants.Production)
	viper.Set(constants.Production, value)
	t.Cleanup(func() { viper.Set(constants.Production, previous) })
}

func snowflakes(emojis []entities.Emoji) map[string]string {
	result := make(map[string]string, len(emojis))
	for _, emoji := range emojis {
		result[emoji.ID] = emoji.Snowflake
	}
	return result
}

func TestGetEmojis_ProductionUsesProductionSnowflakes(t *testing.T) {
	setProduction(t, true)

	got, err := New(newConn(t)).GetEmojis()

	require.NoError(t, err)
	assert.ElementsMatch(t, []entities.Emoji{
		{ID: "dofus", Type: constants.EmojiTypeItem, DiscordName: "dofus_emerald", Name: "Emerald", Snowflake: "prod-1"},
		{ID: "cape", Type: constants.EmojiTypeEquipment, DiscordName: "cape", Name: "Cape", Snowflake: "prod-2"},
	}, got)
}

func TestGetEmojis_NonProductionSwapsInDevSnowflakes(t *testing.T) {
	setProduction(t, false)

	got, err := New(newConn(t)).GetEmojis()

	require.NoError(t, err)
	assert.ElementsMatch(t, []entities.Emoji{
		{ID: "dofus", Type: constants.EmojiTypeItem, DiscordName: "dofus_emerald", Name: "Emerald", Snowflake: "dev-1"},
		{ID: "cape", Type: constants.EmojiTypeEquipment, DiscordName: "cape", Name: "Cape", Snowflake: "dev-2"},
	}, got)
}

func TestGetEmojis_UnsetProductionFlagBehavesAsDev(t *testing.T) {
	setProduction(t, nil)

	got, err := New(newConn(t)).GetEmojis()

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"dofus": "dev-1", "cape": "dev-2"}, snowflakes(got))
}

func TestGetEmojis_EmptyTable(t *testing.T) {
	setProduction(t, true)
	conn := newConn(t)
	require.NoError(t, conn.DB.Exec("DELETE FROM emojis").Error)

	got, err := New(conn).GetEmojis()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetEmojis_DevModeFailsWithoutDevSnowflakeColumn(t *testing.T) {
	setProduction(t, false)
	// AutoMigrate from the entity alone does not create snowflake_dev.
	conn := testutils.NewSQLiteConnection(t, &entities.Emoji{})

	_, err := New(conn).GetEmojis()

	require.Error(t, err)
}

func TestGetEmojis_PropagatesDatabaseError(t *testing.T) {
	for name, production := range map[string]bool{"production": true, "dev": false} {
		t.Run(name, func(t *testing.T) {
			setProduction(t, production)
			conn := newConn(t)
			require.NoError(t, conn.DB.Migrator().DropTable("emojis"))

			got, err := New(conn).GetEmojis()

			require.Error(t, err)
			assert.Empty(t, got)
		})
	}
}
