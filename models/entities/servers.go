package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Server struct {
	Id                  string `gorm:"primaryKey"`
	DofusPortalsId      string `gorm:"unique"`
	DofusEncyclopediaId string `gorm:"unique"`
	Icon                string
	Emoji               string
	Labels              []ServerLabel `gorm:"foreignKey:ServerId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ServerLabel struct {
	Locale   amqp.Language `gorm:"primaryKey"`
	ServerId string        `gorm:"primaryKey"`
	Label    string
}

func (server Server) GetId() string {
	return server.Id
}

func (server Server) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range server.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
