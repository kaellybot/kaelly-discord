package almanaxes

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/almanaxes"

	"github.com/rs/zerolog/log"
)

func New(repository repository.Repository) (*Impl, error) {
	almanaxNews, err := repository.GetAlmanaxNews()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(almanaxNews)).
		Msgf("Almanax News loaded")

	return &Impl{
		almanaxNews: almanaxNews,
		repository:  repository,
	}, nil
}

func (service *Impl) GetAlmanaxNews(locale discordgo.Locale) *entities.AlmanaxNews {
	lg := constants.MapDiscordLocale(locale)

	for _, news := range service.almanaxNews {
		if news.Locale == lg {
			return &news
		}
	}

	return nil
}
