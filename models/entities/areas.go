package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Area struct {
	Id             string      `gorm:"primaryKey"`
	DofusPortalsId string      `gorm:"unique"`
	Labels         []AreaLabel `gorm:"foreignKey:AreaId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type AreaLabel struct {
	Locale amqp.Language `gorm:"primaryKey"`
	AreaId string        `gorm:"primaryKey"`
	Label  string
}

func (area Area) GetId() string {
	return area.Id
}

func (area Area) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range area.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
