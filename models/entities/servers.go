package entities

import amqp "github.com/kaellybot/kaelly-amqp"

//nolint:lll
type Server struct {
	ID                  string `gorm:"primaryKey"`
	DofusPortalsID      string `gorm:"unique"`
	DofusEncyclopediaID string `gorm:"unique"`
	Icon                string
	Emoji               string
	Game                amqp.Game
	Labels              []ServerLabel `gorm:"foreignKey:ServerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ServerLabel struct {
	Locale   amqp.Language `gorm:"primaryKey"`
	ServerID string        `gorm:"primaryKey"`
	Label    string
}

func (server Server) GetID() string {
	return server.ID
}

func (server Server) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range server.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
