// Package testutils holds the shared doubles used by the test suites: an
// in-memory SQLite MySQLConnection, a recording AMQP broker, and a Discord
// session whose REST calls are captured instead of sent. It is only meant to
// be imported from _test.go files.
package testutils

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/glebarez/sqlite"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	di18n "github.com/kaysoro/discordgo-i18n"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// LoadI18n loads the embedded translation bundles, mirroring main.initI18n.
// Tests asserting on translated content must call it (typically in TestMain).
func LoadI18n() {
	di18n.SetDefault(i18n.DefaultLocale)
	for _, language := range i18n.GetLanguages() {
		err := di18n.LoadBundleFS(language.Locale, i18n.Folder, language.TranslationFile)
		if err != nil {
			panic(err)
		}
	}
}

// SQLiteConnection adapts an in-memory SQLite *gorm.DB to the
// databases.MySQLConnection interface. SQLite is dialect-close-enough for the
// repository queries (filters, joins, preloads, upserts); it is not a
// substitute for MySQL on engine-specific behavior.
type SQLiteConnection struct {
	DB *gorm.DB
}

func (c *SQLiteConnection) GetDB() *gorm.DB   { return c.DB }
func (c *SQLiteConnection) IsConnected() bool { return true }
func (c *SQLiteConnection) Run() error        { return nil }
func (c *SQLiteConnection) Shutdown()         {}

// NewSQLiteConnection returns a fresh, isolated in-memory database with the
// given entities migrated.
func NewSQLiteConnection(t *testing.T, models ...any) *SQLiteConnection {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Discard,
	})
	require.NoError(t, err)
	// One connection only: each new connection to ":memory:" is a new database.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if len(models) > 0 {
		require.NoError(t, db.AutoMigrate(models...))
	}
	return &SQLiteConnection{DB: db}
}

// Translations returns every raw variant of a dotted translation key for the
// given locale (keys holding an array are picked at random by i18n.Get, so
// assertions must accept any of them). Placeholders are left unresolved.
func Translations(t *testing.T, locale discordgo.Locale, key string) []string {
	t.Helper()
	var file string
	for _, language := range i18n.GetLanguages() {
		if language.Locale == locale {
			file = language.TranslationFile
		}
	}
	require.NotEmpty(t, file, "unknown locale %v", locale)
	raw, err := i18n.Folder.ReadFile(file)
	require.NoError(t, err)

	var node any
	require.NoError(t, json.Unmarshal(raw, &node))
	for part := range strings.SplitSeq(key, ".") {
		object, ok := node.(map[string]any)
		require.True(t, ok, "key %q not found in %s", key, file)
		node, ok = object[part]
		require.True(t, ok, "key %q not found in %s", key, file)
	}

	switch value := node.(type) {
	case string:
		return []string{value}
	case []any:
		variants := make([]string, 0, len(value))
		for _, variant := range value {
			text, ok := variant.(string)
			require.True(t, ok, "key %q in %s holds a non-string variant", key, file)
			variants = append(variants, text)
		}
		return variants
	default:
		t.Fatalf("key %q in %s is not a translation leaf", key, file)
		return nil
	}
}
