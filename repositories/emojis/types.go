package emojis

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetEmojis() ([]entities.Emoji, error)
}

type Impl struct {
	db databases.MySQLConnection
}
