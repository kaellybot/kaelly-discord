package entities

import "github.com/bwmarrin/discordgo"

type TransportType struct {
	Id             string               `gorm:"primaryKey"`
	DofusPortalsId string               `gorm:"unique"`
	Labels         []TransportTypeLabel `gorm:"foreignKey:TransportTypeId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type TransportTypeLabel struct {
	Locale          discordgo.Locale `gorm:"primaryKey"`
	TransportTypeId string           `gorm:"primaryKey"`
	Label           string
}

func (transportType TransportType) GetId() string {
	return transportType.Id
}

func (transportType TransportType) GetLabels() map[discordgo.Locale]string {
	labels := make(map[discordgo.Locale]string)

	for _, label := range transportType.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
