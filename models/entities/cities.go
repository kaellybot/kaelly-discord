package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type City struct {
	ID     string `gorm:"primaryKey"`
	Icon   string
	Emoji  string
	Type   constants.CityType
	Color  int
	Labels []CityLabel `gorm:"foreignKey:CityID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type CityLabel struct {
	Locale amqp.Language `gorm:"primaryKey"`
	CityID string        `gorm:"primaryKey"`
	Label  string
}

func (city City) GetID() string {
	return city.ID
}

func (city City) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range city.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}

func GetNeutralCity() City {
	return City{
		Color: constants.NeutralCityColor,
		Icon:  "https://raw.githubusercontent.com/kaellybot/kaelly-cdn/refs/heads/main/kaellybot/cities/neutral.webp",
	}
}
