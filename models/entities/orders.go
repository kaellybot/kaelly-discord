package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Order struct {
	Id     string `gorm:"primaryKey"`
	Icon   string
	Emoji  string
	Color  int
	Labels []OrderLabel `gorm:"foreignKey:OrderId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type OrderLabel struct {
	Locale  amqp.Language `gorm:"primaryKey"`
	OrderId string        `gorm:"primaryKey"`
	Label   string
}

func (order Order) GetId() string {
	return order.Id
}

func (order Order) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range order.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
