package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type AlmanaxNews struct {
	Locale        amqp.Language `gorm:"primaryKey"`
	Game          amqp.Game     `gorm:"primaryKey"`
	NewsChannelID string        `gorm:"type:varchar(100)"`
}
