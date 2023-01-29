package entities

import "github.com/bwmarrin/discordgo"

type Order struct {
	Id     string `gorm:"primaryKey"`
	Icon   string
	Emoji  string
	Color  int
	Labels []OrderLabel `gorm:"foreignKey:OrderId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type OrderLabel struct {
	Locale  discordgo.Locale `gorm:"primaryKey"`
	OrderId string           `gorm:"primaryKey"`
	Label   string
}

func (order Order) GetId() string {
	return order.Id
}

func (order Order) GetLabels() map[discordgo.Locale]string {
	labels := make(map[discordgo.Locale]string)

	for _, label := range order.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
