package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type TwitterAccount struct {
	ID   string `gorm:"primaryKey"`
	Name string
	Game amqp.Game
}

func (twittterAccount TwitterAccount) GetID() string {
	return twittterAccount.ID
}

func (twittterAccount TwitterAccount) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, language := range constants.GetLanguages() {
		labels[language.AMQPLocale] = twittterAccount.Name
	}

	return labels
}
