package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Job struct {
	Id     string `gorm:"primaryKey"`
	Icon   string
	Color  int
	Labels []JobLabel `gorm:"foreignKey:JobId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type JobLabel struct {
	Locale amqp.Language `gorm:"primaryKey"`
	JobId  string        `gorm:"primaryKey"`
	Label  string
}

func (job Job) GetId() string {
	return job.Id
}

func (job Job) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range job.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
