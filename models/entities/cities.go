package entities

import "github.com/bwmarrin/discordgo"

type City struct {
	Id     string `gorm:"primaryKey"`
	Icon   string
	Emoji  string
	Color  int
	Labels []CityLabel `gorm:"foreignKey:CityId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type CityLabel struct {
	Locale discordgo.Locale `gorm:"primaryKey"`
	CityId string           `gorm:"primaryKey"`
	Label  string
}

func (city City) GetId() string {
	return city.Id
}

func (city City) GetLabels() map[discordgo.Locale]string {
	labels := make(map[discordgo.Locale]string)

	for _, label := range city.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
