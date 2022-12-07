package entities

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type Server struct {
	Id                  string `gorm:"primaryKey"`
	DofusPortalsId      string `gorm:"unique"`
	DofusEncyclopediaId string `gorm:"unique"`
	Icon                string
	Game                constants.AnkamaGame
}
