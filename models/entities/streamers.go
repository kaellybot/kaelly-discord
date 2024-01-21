package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type Streamer struct {
	ID     string `gorm:"primaryKey"`
	Name   string
	Locale amqp.Language
}

func (streamer Streamer) GetID() string {
	return streamer.ID
}

func (streamer Streamer) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, language := range constants.GetLanguages() {
		labels[language.AMQPLocale] = streamer.Name
	}

	return labels
}
