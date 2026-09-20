package guilds

import (
	"errors"
	"testing"

	"github.com/kaellybot/kaelly-discord/models/entities"
	guildRepo "github.com/kaellybot/kaelly-discord/repositories/guilds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeGuildRepo struct {
	guildRepo.Repository
	exists       bool
	server       entities.Server
	found        bool
	err          error
	gotGuildID   string
	gotChannelID string
}

func (repo *fakeGuildRepo) Exists(guildID string) (bool, error) {
	repo.gotGuildID = guildID
	return repo.exists, repo.err
}

func (repo *fakeGuildRepo) GetServer(guildID, channelID string) (entities.Server, bool, error) {
	repo.gotGuildID, repo.gotChannelID = guildID, channelID
	return repo.server, repo.found, repo.err
}

func TestExists_DelegatesToRepository(t *testing.T) {
	repo := &fakeGuildRepo{exists: true}
	exists, err := New(repo).Exists("g1")
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "g1", repo.gotGuildID)

	boom := errors.New("db down")
	_, err = New(&fakeGuildRepo{err: boom}).Exists("g1")
	assert.ErrorIs(t, err, boom)
}

func TestGetServer_PassesGuildAndChannelInOrder(t *testing.T) {
	repo := &fakeGuildRepo{server: entities.Server{ID: "dakal"}, found: true}
	server, found, err := New(repo).GetServer("g1", "c1")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "dakal", server.ID)
	assert.Equal(t, "g1", repo.gotGuildID)
	assert.Equal(t, "c1", repo.gotChannelID)

	boom := errors.New("db down")
	_, _, err = New(&fakeGuildRepo{err: boom}).GetServer("g1", "c1")
	assert.ErrorIs(t, err, boom)
}
