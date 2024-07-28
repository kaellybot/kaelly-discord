package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Job struct {
	ID     string `gorm:"primaryKey"`
	Icon   string
	Color  int
	Game   amqp.Game
	Labels []JobLabel `gorm:"foreignKey:JobID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type JobLabel struct {
	Locale amqp.Language `gorm:"primaryKey"`
	JobID  string        `gorm:"primaryKey"`
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
