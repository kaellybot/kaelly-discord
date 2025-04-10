package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Order struct {
	ID     string       `gorm:"primaryKey"`
	Game   amqp.Game    `gorm:"primaryKey"`
	Labels []OrderLabel `gorm:"foreignKey:OrderID,Game;references:ID,Game;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type OrderLabel struct {
	OrderID string        `gorm:"primaryKey"`
	Game    amqp.Game     `gorm:"primaryKey"`
	Locale  amqp.Language `gorm:"primaryKey"`
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
