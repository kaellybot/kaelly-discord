package feeds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type FeedRepository interface {
	GetFeedTypes() ([]entities.FeedType, error)
}

type FeedRepositoryImpl struct {
	db databases.MySQLConnection
}

func New(db databases.MySQLConnection) *FeedRepositoryImpl {
	return &FeedRepositoryImpl{db: db}
}

func (repo *FeedRepositoryImpl) GetFeedTypes() ([]entities.FeedType, error) {
	var feedTypes []entities.FeedType
	response := repo.db.GetDB().Model(&entities.FeedType{}).Preload("Labels").Find(&feedTypes)
	return feedTypes, response.Error
}
