package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type SubArea struct {
	ID             string         `gorm:"primaryKey"`
	DofusPortalsID string         `gorm:"unique"`
	Labels         []SubAreaLabel `gorm:"foreignKey:SubAreaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type SubAreaLabel struct {
	Locale    amqp.Language `gorm:"primaryKey"`
	SubAreaID string        `gorm:"primaryKey"`
	Label     string
}

func (subArea SubArea) GetID() string {
	return subArea.ID
}

func (subArea SubArea) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range subArea.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
