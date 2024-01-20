package videasts

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetVideasts() ([]entities.Videast, error) {
	var videasts []entities.Videast
	response := repo.db.GetDB().Model(&entities.Videast{}).Find(&videasts)
	return videasts, response.Error
}
