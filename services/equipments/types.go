package equipments

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/weapons"
)

type Service interface {
	GetWeaponAreaEffect(id string) (entities.WeaponAreaEffect, bool)
}

type Impl struct {
	weaponAreaEffects    map[string]entities.WeaponAreaEffect
	weaponAreaEffectRepo weapons.Repository
}
