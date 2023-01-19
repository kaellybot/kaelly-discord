package entities

import (
	"github.com/bwmarrin/discordgo"
)

type Server struct {
	Id                  string `gorm:"primaryKey"`
	DofusPortalsId      string `gorm:"unique"`
	DofusEncyclopediaId string `gorm:"unique"`
	Icon                string
	Emoji               string
	Labels              []ServerLabel `gorm:"foreignKey:ServerId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ServerLabel struct {
	Locale   discordgo.Locale `gorm:"primaryKey"`
	ServerId string           `gorm:"primaryKey"`
	Label    string
}

func (server Server) GetId() string {
	return server.Id
}

func (server Server) GetLabels() map[discordgo.Locale]string {
	labels := make(map[discordgo.Locale]string)

	for _, label := range server.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
