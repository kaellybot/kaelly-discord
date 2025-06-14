package equipments

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/repositories/weapons"
	"github.com/rs/zerolog/log"
)

func New(weaponRepo weapons.Repository) (*Impl, error) {
	effectsEntities, err := weaponRepo.GetWeaponAreaEffects()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(effectsEntities)).
		Msg("Weapon area effects loaded")

	weaponAreaEffects := make(map[string]entities.WeaponAreaEffect)
	for _, weaponAreaEffect := range effectsEntities {
		weaponAreaEffects[weaponAreaEffect.ID] = weaponAreaEffect
	}

	return &Impl{
		weaponAreaEffects:    weaponAreaEffects,
		weaponAreaEffectRepo: weaponRepo,
	}, nil
}

func (service *Impl) GetWeaponAreaEffect(id string) (entities.WeaponAreaEffect, bool) {
	weaponAreaEffect, found := service.weaponAreaEffects[id]
	return weaponAreaEffect, found
}
