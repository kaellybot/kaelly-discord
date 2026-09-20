package jobs

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
	return testutils.NewSQLiteConnection(t, &entities.Job{}, &entities.JobLabel{})
}

func TestGetJobs_FiltersByGameAndPreloadsLabelsOfSameGame(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&[]entities.Job{
		{ID: "alchemist", Icon: "alch.webp", Color: 10, Game: amqp.Game_DOFUS_GAME},
		{ID: "miner", Icon: "miner.webp", Color: 20, Game: amqp.Game_DOFUS_GAME},
		{ID: "alchemist", Icon: "touch.webp", Color: 30, Game: amqp.Game_DOFUS_TOUCH},
	}).Error)
	require.NoError(t, conn.DB.Create(&[]entities.JobLabel{
		{JobID: "alchemist", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, Label: "Alchimiste"},
		{JobID: "alchemist", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_EN, Label: "Alchemist"},
		{JobID: "alchemist", Game: amqp.Game_DOFUS_TOUCH, Locale: amqp.Language_FR, Label: "Alchimiste Touch"},
	}).Error)

	got, err := New(conn).GetJobs()

	require.NoError(t, err)
	require.Len(t, got, 2)
	byID := map[string]entities.Job{}
	for _, job := range got {
		assert.Equal(t, amqp.Game_DOFUS_GAME, job.Game)
		byID[job.ID] = job
	}
	assert.Equal(t, "alch.webp", byID["alchemist"].Icon)
	assert.Equal(t, 10, byID["alchemist"].Color)
	assert.ElementsMatch(t, []entities.JobLabel{
		{JobID: "alchemist", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_FR, Label: "Alchimiste"},
		{JobID: "alchemist", Game: amqp.Game_DOFUS_GAME, Locale: amqp.Language_EN, Label: "Alchemist"},
	}, byID["alchemist"].Labels)
	assert.Equal(t, "miner.webp", byID["miner"].Icon)
	assert.Empty(t, byID["miner"].Labels)
}

func TestGetJobs_EmptyWhenNoJobForCurrentGame(t *testing.T) {
	conn := newConn(t)
	require.NoError(t, conn.DB.Create(&entities.Job{ID: "miner", Game: amqp.Game_DOFUS_RETRO}).Error)

	got, err := New(conn).GetJobs()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetJobs_PropagatesErrors(t *testing.T) {
	for _, table := range []string{"jobs", "job_labels"} {
		t.Run(table+" missing", func(t *testing.T) {
			conn := newConn(t)
			require.NoError(t, conn.DB.Create(&entities.Job{ID: "miner", Game: amqp.Game_DOFUS_GAME}).Error)
			require.NoError(t, conn.DB.Migrator().DropTable(table))

			_, err := New(conn).GetJobs()

			require.Error(t, err)
		})
	}
}
