package entities

import "github.com/bwmarrin/discordgo"

type Area struct {
	Id             string      `gorm:"primaryKey"`
	DofusPortalsId string      `gorm:"unique"`
	Labels         []AreaLabel `gorm:"foreignKey:AreaId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type AreaLabel struct {
	Locale discordgo.Locale `gorm:"primaryKey"`
	AreaId string           `gorm:"primaryKey"`
	Label  string
}

func (area Area) GetId() string {
	return area.Id
}

func (area Area) GetLabels() map[discordgo.Locale]string {
	labels := make(map[discordgo.Locale]string)

	for _, label := range area.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
