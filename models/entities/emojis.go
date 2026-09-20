package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type Emoji struct {
	ID          string              `gorm:"primaryKey"`
	Type        constants.EmojiType `gorm:"primaryKey"`
	DiscordName string
	Name        string
	// Snowflake is not a column of `emojis`: it is projected from the
	// emoji_snowflakes row matching this bot's game and environment.
	Snowflake string `gorm:"->;-:migration"`
}

// EmojiSnowflake holds one application's snowflake for an emoji. An application
// emoji only renders for the application that owns it, and there is one
// application per game and per environment.
type EmojiSnowflake struct {
	EmojiID    string              `gorm:"primaryKey"`
	EmojiType  constants.EmojiType `gorm:"primaryKey"`
	Game       amqp.Game           `gorm:"primaryKey"`
	Production bool                `gorm:"primaryKey"`
	Snowflake  string
}
