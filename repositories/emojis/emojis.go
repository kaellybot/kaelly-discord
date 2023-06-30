package emojis

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetEmojis() ([]entities.Emoji, error) {
	var emojis []entities.Emoji
	response := repo.db.GetDB().Model(&entities.Emoji{}).Find(&emojis)
	return emojis, response.Error
}
