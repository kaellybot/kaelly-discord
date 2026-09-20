package characteristics

import (
	"errors"
	"testing"

	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/characteristics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCharacteristicRepo struct {
	repository.Repository
	characteristics []entities.Characteristic
	err             error
}

func (repo *fakeCharacteristicRepo) GetCharacteristics() ([]entities.Characteristic, error) {
	return repo.characteristics, repo.err
}

func TestNew_PropagatesRepositoryError(t *testing.T) {
	boom := errors.New("db down")
	service, err := New(&fakeCharacteristicRepo{err: boom})
	require.ErrorIs(t, err, boom)
	assert.Nil(t, service)
}

func TestGetCharacteristic_IndexesByID(t *testing.T) {
	service, err := New(&fakeCharacteristicRepo{characteristics: []entities.Characteristic{
		{ID: "vitality", SortOrder: 1, Regexes: []entities.RegexCharacteristic{{Expression: "vita"}}},
		{ID: "wisdom", SortOrder: 2},
	}})
	require.NoError(t, err)

	vitality, found := service.GetCharacteristic("vitality")
	require.True(t, found)
	assert.Equal(t, 1, vitality.SortOrder)
	assert.Equal(t, "vita", vitality.Regexes[0].Expression, "nested regexes are kept")

	wisdom, found := service.GetCharacteristic("wisdom")
	require.True(t, found)
	assert.Equal(t, 2, wisdom.SortOrder)
}

func TestGetCharacteristic_UnknownID_NotFound(t *testing.T) {
	service, err := New(&fakeCharacteristicRepo{})
	require.NoError(t, err)

	characteristic, found := service.GetCharacteristic("vitality")
	assert.False(t, found)
	assert.Zero(t, characteristic)
}

func TestGetCharacteristic_DuplicateIDs_LastOneWins(t *testing.T) {
	service, err := New(&fakeCharacteristicRepo{characteristics: []entities.Characteristic{
		{ID: "vitality", SortOrder: 1},
		{ID: "vitality", SortOrder: 9},
	}})
	require.NoError(t, err)

	vitality, found := service.GetCharacteristic("vitality")
	require.True(t, found)
	assert.Equal(t, 9, vitality.SortOrder)
}
