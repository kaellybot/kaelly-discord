package emojis

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/spf13/viper"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

// GetEmojis loads the snowflakes of this bot's own Discord application. An
// application emoji only renders for the application that owns it, and there is one
// application per game and per environment, so (game, production) selects the row.
//
// The join is a LEFT one on purpose: an emoji with no snowflake for this
// application keeps an empty one, and the caller falls back on its unicode name
// rather than rendering another application's snowflake.
func (repo *Impl) GetEmojis() ([]entities.Emoji, error) {
	var emojis []entities.Emoji
	response := repo.db.GetDB().
		Model(&entities.Emoji{}).
		Select("emojis.id, emojis.type, emojis.name, emojis.discord_name, "+
			"COALESCE(emoji_snowflakes.snowflake, '') AS snowflake").
		Joins("LEFT JOIN emoji_snowflakes "+
			"ON emoji_snowflakes.emoji_id = emojis.id "+
			"AND emoji_snowflakes.emoji_type = emojis.type "+
			"AND emoji_snowflakes.game = ? AND emoji_snowflakes.production = ?",
			constants.GetGame().AMQPGame, viper.GetBool(constants.Production)).
		Find(&emojis)
	return emojis, response.Error
}
