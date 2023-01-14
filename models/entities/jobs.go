package entities

import "github.com/bwmarrin/discordgo"

type Job struct {
	Id     string `gorm:"primaryKey"`
	Icon   string
	Color  int
	Labels []JobLabel `gorm:"foreignKey:JobId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type JobLabel struct {
	Locale discordgo.Locale `gorm:"primaryKey"`
	JobId  string           `gorm:"primaryKey"`
	Label  string
}

func (job Job) GetId() string {
	return job.Id
}

func (job Job) GetLabels() map[discordgo.Locale]string {
	labels := make(map[discordgo.Locale]string)

	for _, label := range job.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
