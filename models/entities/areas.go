package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Area struct {
	ID             string      `gorm:"primaryKey"`
	DofusPortalsID string      `gorm:"unique"`
	Labels         []AreaLabel `gorm:"foreignKey:AreaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type AreaLabel struct {
	Locale amqp.Language `gorm:"primaryKey"`
	AreaID string        `gorm:"primaryKey"`
	Label  string
}

func (area Area) GetID() string {
	return area.ID
}

func (area Area) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range area.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
