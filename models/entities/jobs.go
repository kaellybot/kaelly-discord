package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Job struct {
	ID     string `gorm:"primaryKey"`
	Icon   string
	Color  int
	Game   amqp.Game  `gorm:"primaryKey"`
	Labels []JobLabel `gorm:"foreignKey:JobID,Game;references:ID,Game;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type JobLabel struct {
	JobID  string        `gorm:"primaryKey"`
	Game   amqp.Game     `gorm:"primaryKey"`
	Locale amqp.Language `gorm:"primaryKey"`
	Label  string
}

func (job Job) GetID() string {
	return job.ID
}

func (job Job) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range job.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
