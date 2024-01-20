package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type Videast struct {
	ID     string `gorm:"primaryKey"`
	Name   string
	Locale amqp.Language
}

func (videast Videast) GetID() string {
	return videast.ID
}

func (videast Videast) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, language := range constants.GetLanguages() {
		labels[language.AMQPLocale] = videast.Name
	}

	return labels
}
