package equipments

import (
	"errors"
	"testing"

	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/weapons"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeWeaponRepo struct {
	weapons.Repository
	effects []entities.WeaponAreaEffect
	err     error
}

func (repo *fakeWeaponRepo) GetWeaponAreaEffects() ([]entities.WeaponAreaEffect, error) {
	return repo.effects, repo.err
}

func TestNew_PropagatesRepositoryError(t *testing.T) {
	boom := errors.New("db down")
	service, err := New(&fakeWeaponRepo{err: boom})
	require.ErrorIs(t, err, boom)
	assert.Nil(t, service)
}

func TestGetWeaponAreaEffect_IndexesByID(t *testing.T) {
	service, err := New(&fakeWeaponRepo{effects: []entities.WeaponAreaEffect{
		{ID: "cross", Order: 1},
		{ID: "line", Order: 2},
	}})
	require.NoError(t, err)

	cross, found := service.GetWeaponAreaEffect("cross")
	require.True(t, found)
	assert.Equal(t, 1, cross.Order)

	line, found := service.GetWeaponAreaEffect("line")
	require.True(t, found)
	assert.Equal(t, 2, line.Order)

	missing, found := service.GetWeaponAreaEffect("circle")
	assert.False(t, found)
	assert.Zero(t, missing)
}
