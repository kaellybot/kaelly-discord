package emojis

import (
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// otherGame stands for the other bot's fork: its snowflakes must never be loaded.
func otherGame() amqp.Game {
	if constants.GetGame().AMQPGame == amqp.Game_DOFUS_TOUCH {
		return amqp.Game_DOFUS_GAME
	}
	return amqp.Game_DOFUS_TOUCH
}

// The same two emojis exist in both applications and both environments, with a
// different snowflake each time: only one of the four may ever be loaded.
func newConn(t *testing.T) *testutils.SQLiteConnection {
	t.Helper()
	conn := testutils.NewSQLiteConnection(t, &entities.Emoji{}, &entities.EmojiSnowflake{})
	ours := constants.GetGame().AMQPGame

	require.NoError(t, conn.DB.Create(&[]entities.Emoji{
		{ID: "dofus", Type: constants.EmojiTypeItem, DiscordName: "dofus_emerald", Name: "Emerald"},
		{ID: "cape", Type: constants.EmojiTypeEquipment, DiscordName: "cape", Name: "Cape"},
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.EmojiSnowflake{
		{EmojiID: "dofus", EmojiType: constants.EmojiTypeItem, Game: ours, Production: true, Snowflake: "prod-1"},
		{EmojiID: "dofus", EmojiType: constants.EmojiTypeItem, Game: ours, Production: false, Snowflake: "dev-1"},
		{EmojiID: "cape", EmojiType: constants.EmojiTypeEquipment, Game: ours, Production: true, Snowflake: "prod-2"},
		{EmojiID: "cape", EmojiType: constants.EmojiTypeEquipment, Game: ours, Production: false, Snowflake: "dev-2"},
		{EmojiID: "dofus", EmojiType: constants.EmojiTypeItem,
			Game: otherGame(), Production: true, Snowflake: "theirs-1"},
		{EmojiID: "cape", EmojiType: constants.EmojiTypeEquipment,
			Game: otherGame(), Production: false, Snowflake: "theirs-2"},
	}).Error)
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

func TestGetEmojis_NonProductionUsesDevSnowflakes(t *testing.T) {
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

// Rendering another application's snowflake shows a broken emoji, so an emoji
// with no row for this application keeps an empty one and the caller falls back
// on its unicode name.
func TestGetEmojis_NeverLoadsAnotherApplicationsSnowflake(t *testing.T) {
	setProduction(t, true)
	conn := testutils.NewSQLiteConnection(t, &entities.Emoji{}, &entities.EmojiSnowflake{})
	require.NoError(t, conn.DB.Create(&entities.Emoji{
		ID: "kama", Type: constants.EmojiTypeMisc, DiscordName: "kama", Name: "💰",
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.EmojiSnowflake{
		// Only the other game, and only the other environment, have a snowflake.
		{EmojiID: "kama", EmojiType: constants.EmojiTypeMisc,
			Game: otherGame(), Production: true, Snowflake: "theirs"},
		{EmojiID: "kama", EmojiType: constants.EmojiTypeMisc,
			Game: constants.GetGame().AMQPGame, Production: false, Snowflake: "ours-dev"},
	}).Error)

	got, err := New(conn).GetEmojis()

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Empty(t, got[0].Snowflake, "no snowflake for this game and environment")
	assert.Equal(t, "💰", got[0].Name, "the unicode fallback is still available")
}

// An emoji with no snowflake row at all is still returned, so its unicode name
// can be rendered: the join must be a LEFT one.
func TestGetEmojis_EmojiWithoutAnySnowflakeIsStillReturned(t *testing.T) {
	setProduction(t, true)
	conn := testutils.NewSQLiteConnection(t, &entities.Emoji{}, &entities.EmojiSnowflake{})
	require.NoError(t, conn.DB.Create(&entities.Emoji{
		ID: "kama", Type: constants.EmojiTypeMisc, Name: "💰",
	}).Error)

	got, err := New(conn).GetEmojis()

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Empty(t, got[0].Snowflake)
}

func TestGetEmojis_EmptyTable(t *testing.T) {
	setProduction(t, true)
	conn := newConn(t)
	require.NoError(t, conn.DB.Exec("DELETE FROM emojis").Error)

	got, err := New(conn).GetEmojis()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetEmojis_PropagatesDatabaseError(t *testing.T) {
	for name, table := range map[string]string{
		"emojis missing":           "emojis",
		"emoji_snowflakes missing": "emoji_snowflakes",
	} {
		t.Run(name, func(t *testing.T) {
			setProduction(t, true)
			conn := newConn(t)
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			got, err := New(conn).GetEmojis()

			require.Error(t, err)
			assert.Empty(t, got)
		})
	}
}
