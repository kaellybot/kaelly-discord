package almanaxes

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/almanaxes"
)

type Service interface {
	GetAlmanaxNews(locale discordgo.Locale) *entities.AlmanaxNews
}

type Impl struct {
	almanaxNews []entities.AlmanaxNews
	repository  repository.Repository
}
