package weapons

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetWeaponAreaEffects() ([]entities.WeaponAreaEffect, error) {
	var weaponAreaEffects []entities.WeaponAreaEffect
	response := repo.db.GetDB().
		Model(&entities.WeaponAreaEffect{}).
		Preload("Labels").
		Find(&weaponAreaEffects)
	return weaponAreaEffects, response.Error
}
