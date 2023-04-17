package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type SubArea struct {
	Id             string         `gorm:"primaryKey"`
	DofusPortalsId string         `gorm:"unique"`
	Labels         []SubAreaLabel `gorm:"foreignKey:SubAreaId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type SubAreaLabel struct {
	Locale    amqp.Language `gorm:"primaryKey"`
	SubAreaId string        `gorm:"primaryKey"`
	Label     string
}

func (subArea SubArea) GetId() string {
	return subArea.Id
}

func (subArea SubArea) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range subArea.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
