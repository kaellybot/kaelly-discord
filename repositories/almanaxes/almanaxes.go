package almanaxes

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetAlmanaxNews() ([]entities.AlmanaxNews, error) {
	var almanaxNews []entities.AlmanaxNews
	response := repo.db.GetDB().
		Model(&entities.AlmanaxNews{}).
		Where("game = ?", constants.GetGame().AMQPGame).
		Find(&almanaxNews)
	return almanaxNews, response.Error
}
