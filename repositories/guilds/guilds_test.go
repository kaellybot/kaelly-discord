package guilds

import (
	"testing"

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
	require.NoError(t, conn.DB.Exec(
		"CREATE TABLE guilds (id TEXT PRIMARY KEY, server_id TEXT NULL)").Error)
	require.NoError(t, conn.DB.Exec(`CREATE TABLE channel_servers (
		guild_id TEXT, channel_id TEXT, server_id TEXT,
		PRIMARY KEY (guild_id, channel_id))`).Error)
	return conn
}

func exec(t *testing.T, conn *testutils.SQLiteConnection, sql string) {
	t.Helper()
	require.NoError(t, conn.DB.Exec(sql).Error)
}

func TestExists_TrueOnlyForKnownGuild(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, "INSERT INTO guilds (id) VALUES ('g1'), ('g2')")
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
	require.NoError(t, conn.DB.Create(&entities.Guild{ID: "g1"}).Error)

	exists, err := New(conn).Exists("g1")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExists_FalseOnEmptyTable(t *testing.T) {
	exists, err := New(newConn(t)).Exists("g1")

	require.NoError(t, err)
	assert.False(t, exists)
}

// Where(&Guild{ID: ""}) drops the zero-valued ID from the condition, so an
// empty guild ID matches any guild. Callers only pass real guild IDs today;
// this documents the current behavior rather than endorsing it.
func TestExists_EmptyGuildIDMatchesAnyGuild(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, "INSERT INTO guilds (id) VALUES ('g1')")

	exists, err := New(conn).Exists("")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExists_PropagatesDatabaseError(t *testing.T) {
	conn := testutils.NewSQLiteConnection(t)

	exists, err := New(conn).Exists("g1")

	require.Error(t, err)
	assert.False(t, exists)
}

func TestGetServer_ResolutionOrder(t *testing.T) {
	conn := newConn(t)
	exec(t, conn, `INSERT INTO guilds (id, server_id) VALUES
		('with-default', 'guild-server'),
		('no-default', NULL),
		('empty-default', '')`)
	exec(t, conn, `INSERT INTO channel_servers (guild_id, channel_id, server_id) VALUES
		('with-default', 'override', 'channel-server'),
		('no-default', 'override', 'channel-server'),
		('other-guild', 'shared', 'foreign-server'),
		('with-default', 'null-override', NULL)`)
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
			assert.Equal(t, entities.Server{ID: tc.wantID}, server)
		})
	}
}

func TestGetServer_PropagatesErrors(t *testing.T) {
	for _, table := range []string{"guilds", "channel_servers"} {
		t.Run(table+" missing", func(t *testing.T) {
			conn := newConn(t)
			exec(t, conn, "INSERT INTO guilds (id, server_id) VALUES ('g1', 's1')")
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			server, found, err := New(conn).GetServer("g1", "c1")

			require.Error(t, err)
			assert.False(t, found)
			assert.Equal(t, entities.Server{}, server)
		})
	}
}
