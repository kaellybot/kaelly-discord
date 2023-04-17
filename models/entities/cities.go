package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type City struct {
	Id     string `gorm:"primaryKey"`
	Icon   string
	Emoji  string
	Color  int
	Labels []CityLabel `gorm:"foreignKey:CityId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type CityLabel struct {
	Locale amqp.Language `gorm:"primaryKey"`
	CityId string        `gorm:"primaryKey"`
	Label  string
}

func (city City) GetId() string {
	return city.Id
}

func (city City) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range city.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
