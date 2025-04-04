package feeds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/feeds"
	"golang.org/x/text/transform"
)

type Service interface {
	GetFeedTypes() []entities.FeedType
	GetFeedType(ID string) *entities.FeedType
	FindFeedTypes(name string, locale discordgo.Locale, limit int) []entities.FeedType
}

type Impl struct {
	transformer transform.Transformer
	feedTypes   []entities.FeedType
	repository  repository.Repository
}
