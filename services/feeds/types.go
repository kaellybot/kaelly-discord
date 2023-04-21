package feeds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/feeds"
	"golang.org/x/text/transform"
)

type FeedService interface {
	GetFeedTypes() []entities.FeedType
	FindFeedTypes(name string, locale discordgo.Locale) []entities.FeedType
}

type FeedServiceImpl struct {
	transformer transform.Transformer
	feedTypes   []entities.FeedType
	repository  repository.FeedRepository
}
