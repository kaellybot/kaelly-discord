package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/i18n"
)

// TwitterAccount is identified by (id, game): each game has its own accounts and
// its own news channel to follow.
type TwitterAccount struct {
	ID            string    `gorm:"primaryKey"`
	Game          amqp.Game `gorm:"primaryKey"`
	Name          string
	NewsChannelID string
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
