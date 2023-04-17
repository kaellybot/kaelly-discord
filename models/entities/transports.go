package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
)

type TransportType struct {
	Id             string `gorm:"primaryKey"`
	DofusPortalsId string `gorm:"unique"`
	Emoji          string
	Labels         []TransportTypeLabel `gorm:"foreignKey:TransportTypeId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type TransportTypeLabel struct {
	Locale          amqp.Language `gorm:"primaryKey"`
	TransportTypeId string        `gorm:"primaryKey"`
	Label           string
}

func (transportType TransportType) GetId() string {
	return transportType.Id
}

func (transportType TransportType) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range transportType.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
