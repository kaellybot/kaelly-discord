package videasts

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/videasts"

	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func New(repository repository.Repository) (*Impl, error) {
	videasts, err := repository.GetVideasts()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(videasts)).
		Msgf("Videasts loaded")

	return &Impl{
		transformer: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		videasts:    videasts,
		repository:  repository,
	}, nil
}

func (service *Impl) GetVideasts() []entities.Videast {
	return service.videasts
}

func (service *Impl) GetVideast(ID string) *entities.Videast {
	for _, videast := range service.videasts {
		if videast.ID == ID {
			return &videast
		}
	}

	return nil
}

func (service *Impl) FindVideasts(name string, locale discordgo.Locale) []entities.Videast {
	videastsFound := make([]entities.Videast, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize videast name, returning empty videast name list")
		return videastsFound
	}

	for _, videast := range service.videasts {
		currentCleanedName, _, errStr := transform.String(service.transformer,
			strings.ToLower(translators.GetEntityLabel(videast, locale)))
		if errStr == nil && strings.HasPrefix(currentCleanedName, cleanedName) {
			videastsFound = append(videastsFound, videast)
		}
	}

	return videastsFound
}
