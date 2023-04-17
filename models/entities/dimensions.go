package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Dimension struct {
	Id             string `gorm:"primaryKey"`
	DofusPortalsId string `gorm:"unique"`
	Icon           string
	Color          int
	Labels         []DimensionLabel `gorm:"foreignKey:DimensionId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type DimensionLabel struct {
	Locale      amqp.Language `gorm:"primaryKey"`
	DimensionId string        `gorm:"primaryKey"`
	Label       string
}

func (dimension Dimension) GetId() string {
	return dimension.Id
}

func (dimension Dimension) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range dimension.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
