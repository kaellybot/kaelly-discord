package entities

import amqp "github.com/kaellybot/kaelly-amqp"

// Server is identified by (id, game): the same server ID may exist in several
// games with different data, so its labels are keyed by game too.
type Server struct {
	ID     string    `gorm:"primaryKey"`
	Game   amqp.Game `gorm:"primaryKey"`
	Icon   string
	Image  string
	Labels []ServerLabel `gorm:"foreignKey:ServerID,Game;references:ID,Game;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ServerLabel struct {
	Locale   amqp.Language `gorm:"primaryKey"`
	ServerID string        `gorm:"primaryKey"`
	Game     amqp.Game     `gorm:"primaryKey"`
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
