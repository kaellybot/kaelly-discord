package guilds

import (
	"fmt"
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// guilds.server_id and channel_servers are not mapped by entities, so the
// schema is created by hand to mirror production.
func newConn(t *testing.T) *testutils.SQLiteConnection {
	t.Helper()
	conn := testutils.NewSQLiteConnection(t)
	require.NoError(t, conn.DB.Exec(`CREATE TABLE guilds (
		id TEXT, game INTEGER, server_id TEXT NULL,
		PRIMARY KEY (id, game))`).Error)
	require.NoError(t, conn.DB.Exec(`CREATE TABLE channel_servers (
		guild_id TEXT, channel_id TEXT, game INTEGER, server_id TEXT,
		PRIMARY KEY (guild_id, channel_id, game))`).Error)
	return conn
}

// thisGame is the game of this bot; otherGame stands for the other bot's fork.
func thisGame() amqp.Game {
	return constants.GetGame().AMQPGame
}

func otherGame() amqp.Game {
	if thisGame() == amqp.Game_DOFUS_TOUCH {
		return amqp.Game_DOFUS_GAME
	}
	return amqp.Game_DOFUS_TOUCH
}

func exec(t *testing.T, conn *testutils.SQLiteConnection, sql string) {
	t.Helper()
	require.NoError(t, conn.DB.Exec(sql).Error)
}

func TestExists_TrueOnlyForKnownGuild(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, fmt.Sprintf("INSERT INTO guilds (id, game) VALUES ('g1', %d), ('g2', %d)",
		thisGame(), thisGame()))
	repo := New(conn)

	cases := map[string]bool{"g1": true, "g2": true, "g3": false}
	for guildID, expected := range cases {
		t.Run(guildID, func(t *testing.T) {
			exists, err := repo.Exists(guildID)

			require.NoError(t, err)
			assert.Equal(t, expected, exists)
		})
	}
}

func TestExists_WorksWithEntityMigratedTable(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t, &entities.Guild{})
	require.NoError(t, conn.DB.Create(&entities.Guild{ID: "g1", Game: thisGame()}).Error)

	exists, err := New(conn).Exists("g1")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExists_FalseOnEmptyTable(t *testing.T) {
	exists, err := New(newConn(t)).Exists("g1")

	require.NoError(t, err)
	assert.False(t, exists)
}

// Where(&Guild{ID: ""}) drops the zero-valued ID from the condition, which
// would make an empty guild ID match any guild: an empty ID is not a guild.
func TestExists_EmptyGuildIDIsNeverAGuild(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, fmt.Sprintf("INSERT INTO guilds (id, game) VALUES ('g1', %d)", thisGame()))

	exists, err := New(conn).Exists("")

	require.NoError(t, err)
	assert.False(t, exists)
}

// The empty ID is rejected before any query, so a broken schema is not reached.
func TestExists_EmptyGuildIDDoesNotQuery(t *testing.T) {
	exists, err := New(testutils.NewSQLiteConnection(t)).Exists("")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestExists_PropagatesDatabaseError(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t)

	exists, err := New(conn).Exists("g1")

	require.Error(t, err)
	assert.False(t, exists)
}

func TestGetServer_ResolutionOrder(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, fmt.Sprintf(`INSERT INTO guilds (id, game, server_id) VALUES
		('with-default', %[1]d, 'guild-server'),
		('no-default', %[1]d, NULL),
		('empty-default', %[1]d, '')`, thisGame()))
	exec(t, conn, fmt.Sprintf(`INSERT INTO channel_servers (guild_id, channel_id, game, server_id) VALUES
		('with-default', 'override', %[1]d, 'channel-server'),
		('no-default', 'override', %[1]d, 'channel-server'),
		('other-guild', 'shared', %[1]d, 'foreign-server'),
		('with-default', 'null-override', %[1]d, NULL)`, thisGame()))
	repo := New(conn)

	cases := []struct {
		name, guildID, channelID string
		wantID                   string
		wantFound                bool
	}{
		{"channel override wins over guild default", "with-default", "override", "channel-server", true},
		{"guild default when channel has no override", "with-default", "plain", "guild-server", true},
		{"guild default when channel override is NULL", "with-default", "null-override", "guild-server", true},
		{"channel override without guild default", "no-default", "override", "channel-server", true},
		{"nothing configured", "no-default", "plain", "", false},
		{"empty guild default counts as none", "empty-default", "plain", "", false},
		{"override of another guild is ignored", "with-default", "shared", "guild-server", true},
		{"unknown guild", "unknown", "override", "", false},
		{"direct message (no guild)", "", "override", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, found, err := repo.GetServer(tc.guildID, tc.channelID)

			require.NoError(t, err)
			assert.Equal(t, tc.wantFound, found)
			want := entities.Server{}
			if tc.wantFound {
				want = entities.Server{ID: tc.wantID, Game: thisGame()}
			}
			assert.Equal(t, want, server)
		})
	}
}

func TestGetServer_PropagatesErrors(t *testing.T) {
	for _, table := range []string{"guilds", "channel_servers"} {
		t.Run(table+" missing", func(t *testing.T) {
			conn := newConn(t)
			exec(t, conn, fmt.Sprintf("INSERT INTO guilds (id, game, server_id) VALUES ('g1', %d, 's1')",
				thisGame()))
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			server, found, err := New(conn).GetServer("g1", "c1")

			require.Error(t, err)
			assert.False(t, found)
			assert.Equal(t, entities.Server{}, server)
		})
	}
}

// A guild registered by the other game's bot is not this bot's guild: treating it
// as known would suppress this bot's create news and stop the configurator from
// ever creating its row.
func TestExists_IgnoresOtherGamesGuild(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, fmt.Sprintf("INSERT INTO guilds (id, game) VALUES ('shared', %d)", otherGame()))

	exists, err := New(conn).Exists("shared")

	require.NoError(t, err)
	assert.False(t, exists)
}

// The same guild can host both bots, each with its own server configuration.
func TestGetServer_IsScopedToThisGame(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, fmt.Sprintf(`INSERT INTO guilds (id, game, server_id) VALUES
		('shared', %d, 'ours'),
		('shared', %d, 'theirs')`, thisGame(), otherGame()))
	exec(t, conn, fmt.Sprintf(`INSERT INTO channel_servers (guild_id, channel_id, game, server_id) VALUES
		('shared', 'chan', %d, 'their-channel-server')`, otherGame()))

	server, found, err := New(conn).GetServer("shared", "chan")

	require.NoError(t, err)
	assert.True(t, found)
	// The other game's channel override must not leak in through the join.
	assert.Equal(t, entities.Server{ID: "ours", Game: thisGame()}, server)
}

// A guild that only the other bot knows resolves to nothing here.
func TestGetServer_IgnoresOtherGamesGuild(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, fmt.Sprintf("INSERT INTO guilds (id, game, server_id) VALUES ('theirs', %d, 's1')",
		otherGame()))

	server, found, err := New(conn).GetServer("theirs", "chan")

	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, entities.Server{}, server)
}
