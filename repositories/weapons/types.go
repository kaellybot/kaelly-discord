package weapons

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetWeaponAreaEffects() ([]entities.WeaponAreaEffect, error)
}

type Impl struct {
	db databases.MySQLConnection
}
