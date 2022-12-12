package entities

import "github.com/bwmarrin/discordgo"

type Dimension struct {
	Id             string `gorm:"primaryKey"`
	DofusPortalsId string `gorm:"unique"`
	Icon           string
	Color          int
	Labels         []DimensionLabel `gorm:"foreignKey:DimensionId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type DimensionLabel struct {
	Locale      discordgo.Locale `gorm:"primaryKey"`
	DimensionId string           `gorm:"primaryKey"`
	Label       string
}

func (dimension Dimension) GetId() string {
	return dimension.Id
}

func (dimension Dimension) GetLabels() map[discordgo.Locale]string {
	labels := make(map[discordgo.Locale]string)

	for _, label := range dimension.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
