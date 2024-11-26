package twitters

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/twitters"
	"golang.org/x/text/transform"
)

type Service interface {
	GetTwitterAccounts() []entities.TwitterAccount
	GetTwitterAccount(ID string) *entities.TwitterAccount
	FindTwitterAccounts(name string, locale discordgo.Locale, limit int) []entities.TwitterAccount
}

type Impl struct {
	transformer     transform.Transformer
	twitterAccounts []entities.TwitterAccount
	repository      repository.Repository
}
