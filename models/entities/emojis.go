package entities

import "github.com/kaellybot/kaelly-discord/models/constants"

type Emoji struct {
	ID          string              `gorm:"primaryKey"`
	Type        constants.EmojiType `gorm:"primaryKey"`
	DiscordName string
	Snowflake   string
	Name        string
}
