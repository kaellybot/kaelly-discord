package feeds

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetFeedTypes() ([]entities.FeedType, error) {
	var feedTypes []entities.FeedType
	response := repo.db.GetDB().
		Model(&entities.FeedType{}).
		Preload("Labels").
		Preload("Sources", "game = ?", constants.GetGame().AMQPGame).
		Find(&feedTypes)
	return feedTypes, response.Error
}
