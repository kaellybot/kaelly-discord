package cities

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetCharacteristics() ([]entities.Characteristic, error) {
	var response *gorm.DB
	var characteristics []entities.Characteristic
	if viper.GetBool(constants.Production) {
		response = repo.db.GetDB().
			Model(entities.Characteristic{}).
			Find(&characteristics)
	} else {
		log.Info().Msgf("Development mode enabled, retrieving characteristic debug emojis")
		response = repo.db.GetDB().
			Model(&entities.Characteristic{}).
			Select("id, emoji_dev AS emoji, sort_order").
			Find(&characteristics)
	}

	return characteristics, response.Error
}
