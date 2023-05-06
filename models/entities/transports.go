package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
)

//nolint:lll
type TransportType struct {
	ID             string `gorm:"primaryKey"`
	DofusPortalsID string `gorm:"unique"`
	Emoji          string
	Labels         []TransportTypeLabel `gorm:"foreignKey:TransportTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type TransportTypeLabel struct {
	Locale          amqp.Language `gorm:"primaryKey"`
	TransportTypeID string        `gorm:"primaryKey"`
	Label           string
}

func (transportType TransportType) GetID() string {
	return transportType.ID
}

func (transportType TransportType) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range transportType.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
