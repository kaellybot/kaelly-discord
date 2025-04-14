package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/i18n"
)

type TwitterAccount struct {
	ID            string `gorm:"primaryKey"`
	Name          string
	NewsChannelID string
	Game          amqp.Game
}

func (twittterAccount TwitterAccount) GetID() string {
	return twittterAccount.ID
}

func (twittterAccount TwitterAccount) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, language := range i18n.GetLanguages() {
		labels[language.AMQPLocale] = twittterAccount.Name
	}

	return labels
}
