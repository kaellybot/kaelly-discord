package feeds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/feeds"
)

type Service interface {
	GetFeedTypes() []entities.FeedType
	GetFeedType(ID string) *entities.FeedType
	FindFeedTypes(name string, locale discordgo.Locale, limit int) []entities.FeedType
}

type Impl struct {
	feedTypes  []entities.FeedType
	repository repository.Repository
}
