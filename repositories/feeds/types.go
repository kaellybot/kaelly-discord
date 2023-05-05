package feeds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetFeedTypes() ([]entities.FeedType, error)
}

type Impl struct {
	db databases.MySQLConnection
}
