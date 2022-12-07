package entities

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type Server struct {
	Id                  string `gorm:"primaryKey"`
	DofusPortalsId      string `gorm:"unique"`
	DofusEncyclopediaId string `gorm:"unique"`
	Icon                string
	Game                constants.AnkamaGame
	Labels              []ServerLabel `gorm:"foreignKey:ServerId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ServerLabel struct {
	Locale   discordgo.Locale `gorm:"primaryKey"`
	ServerId string           `gorm:"primaryKey"`
	Label    string
}
