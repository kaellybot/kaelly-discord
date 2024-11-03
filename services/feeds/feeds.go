package feeds

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/feeds"

	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func New(repository repository.Repository) (*Impl, error) {
	feedTypes, err := repository.GetFeedTypes()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(feedTypes)).
		Msgf("Feed types loaded")

	return &Impl{
		transformer: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		feedTypes:   feedTypes,
		repository:  repository,
	}, nil
}

func (service *Impl) GetFeedTypes() []entities.FeedType {
	return service.feedTypes
}

func (service *Impl) FindFeedTypes(name string, locale discordgo.Locale) []entities.FeedType {
	feedTypesFound := make([]entities.FeedType, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize feed type name, returning empty feed type list")
		return feedTypesFound
	}

	for _, feedType := range service.feedTypes {
		currentCleanedName, _, errStr := transform.String(service.transformer,
			strings.ToLower(translators.GetEntityLabel(feedType, locale)))
		if errStr == nil && strings.Contains(currentCleanedName, cleanedName) {
			if currentCleanedName == cleanedName {
				return []entities.FeedType{feedType}
			}

			feedTypesFound = append(feedTypesFound, feedType)
		}
	}

	return feedTypesFound
}
