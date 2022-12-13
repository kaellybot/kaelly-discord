package entities

import "github.com/bwmarrin/discordgo"

type SubArea struct {
	Id             string         `gorm:"primaryKey"`
	DofusPortalsId string         `gorm:"unique"`
	Labels         []SubAreaLabel `gorm:"foreignKey:SubAreaId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type SubAreaLabel struct {
	Locale    discordgo.Locale `gorm:"primaryKey"`
	SubAreaId string           `gorm:"primaryKey"`
	Label     string
}

func (subArea SubArea) GetId() string {
	return subArea.Id
}

func (subArea SubArea) GetLabels() map[discordgo.Locale]string {
	labels := make(map[discordgo.Locale]string)

	for _, label := range subArea.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
