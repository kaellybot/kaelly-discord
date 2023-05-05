package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Order struct {
	ID     string `gorm:"primaryKey"`
	Icon   string
	Emoji  string
	Color  int
	Labels []OrderLabel `gorm:"foreignKey:OrderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type OrderLabel struct {
	Locale  amqp.Language `gorm:"primaryKey"`
	OrderID string        `gorm:"primaryKey"`
	Label   string
}

func (order Order) GetID() string {
	return order.ID
}

func (order Order) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range order.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
