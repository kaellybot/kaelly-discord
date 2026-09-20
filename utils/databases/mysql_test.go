package databases

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setConfig(t *testing.T, values map[string]any) {
	t.Helper()
	for key, value := range values {
		previous := viper.Get(key)
		t.Cleanup(func() { viper.Set(key, previous) })
		viper.Set(key, value)
	}
}

func TestNew_BuildsDSNFromConfig(t *testing.T) {
	setConfig(t, map[string]any{
		constants.MySQLUser:     "kaelly",
		constants.MySQLPassword: "p@ss:word",
		constants.MySQLURL:      "db.local:3307",
		constants.MySQLDatabase: "kaellybot",
	})

	conn, ok := New().(*mySQLConnection)
	require.True(t, ok)

	assert.Equal(t, "kaelly:p@ss:word@tcp(db.local:3307)/kaellybot?charset=utf8mb4&parseTime=true&loc=UTC",
		conn.dsn, "credentials are inserted verbatim, times are parsed as UTC")
}

func TestNew_DoesNotConnect(t *testing.T) {
	conn := New()

	assert.Nil(t, conn.GetDB())
	assert.False(t, conn.IsConnected(), "no database until Run is called")
}

func TestIsConnected_ReflectsUnderlyingPool(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	conn := &mySQLConnection{db: db}

	assert.True(t, conn.IsConnected())
	assert.Same(t, db, conn.GetDB())

	conn.Shutdown()
	assert.False(t, conn.IsConnected(), "a closed pool no longer answers pings")
	assert.NotPanics(t, conn.Shutdown, "closing twice is harmless")
}

func TestIsConnected_FalseWhenGormHasNoSQLPool(t *testing.T) {
	conn := &mySQLConnection{db: &gorm.DB{Config: &gorm.Config{}}}

	assert.False(t, conn.IsConnected())
}
