package streamers

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/streamers"

	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func New(repository repository.Repository) (*Impl, error) {
	streamers, err := repository.GetStreamers()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(streamers)).
		Msgf("Streamers loaded")

	return &Impl{
		transformer: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		streamers:   streamers,
		repository:  repository,
	}, nil
}

func (service *Impl) GetStreamers() []entities.Streamer {
	return service.streamers
}

func (service *Impl) GetStreamer(id string) *entities.Streamer {
	for _, streamer := range service.streamers {
		if streamer.ID == id {
			return &streamer
		}
	}

	return nil
}

func (service *Impl) FindStreamers(name string, locale discordgo.Locale) []entities.Streamer {
	streamersFound := make([]entities.Streamer, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize streamer name, returning empty streamer name list")
		return streamersFound
	}

	for _, streamer := range service.streamers {
		currentCleanedName, _, errStr := transform.String(service.transformer,
			strings.ToLower(translators.GetEntityLabel(streamer, locale)))
		if errStr == nil && strings.HasPrefix(currentCleanedName, cleanedName) {
			streamersFound = append(streamersFound, streamer)
		}
	}

	return streamersFound
}
