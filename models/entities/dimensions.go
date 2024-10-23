package entities

import amqp "github.com/kaellybot/kaelly-amqp"

//nolint:lll
type Dimension struct {
	ID     string `gorm:"primaryKey"`
	Icon   string
	Color  int
	Labels []DimensionLabel `gorm:"foreignKey:DimensionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type DimensionLabel struct {
	Locale      amqp.Language `gorm:"primaryKey"`
	DimensionID string        `gorm:"primaryKey"`
	Label       string
}

func (dimension Dimension) GetID() string {
	return dimension.ID
}

func (dimension Dimension) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range dimension.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
