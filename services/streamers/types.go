package streamers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/streamers"
	"golang.org/x/text/transform"
)

type Service interface {
	GetStreamers() []entities.Streamer
	GetStreamer(id string) *entities.Streamer
	FindStreamers(name string, locale discordgo.Locale, limit int) []entities.Streamer
}

type Impl struct {
	transformer transform.Transformer
	streamers   []entities.Streamer
	repository  repository.Repository
}
