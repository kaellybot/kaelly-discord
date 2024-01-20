package videasts

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/videasts"
	"golang.org/x/text/transform"
)

type Service interface {
	GetVideasts() []entities.Videast
	FindVideasts(name string, locale discordgo.Locale) []entities.Videast
}

type Impl struct {
	transformer transform.Transformer
	videasts    []entities.Videast
	repository  repository.Repository
}
