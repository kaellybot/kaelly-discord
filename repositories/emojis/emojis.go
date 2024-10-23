package emojis

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetEmojis() ([]entities.Emoji, error) {
	var response *gorm.DB
	var emojis []entities.Emoji
	if viper.GetBool(constants.Production) {
		response = repo.db.GetDB().
			Model(&entities.Emoji{}).
			Find(&emojis)
	} else {
		response = repo.db.GetDB().
			Model(&entities.Emoji{}).
			Select("id, snowflake_dev AS snowflake, type, name").
			Find(&emojis)
	}
	return emojis, response.Error
}
