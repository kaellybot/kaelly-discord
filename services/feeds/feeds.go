package feeds

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/feeds"

	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
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
		feedTypes:  feedTypes,
		repository: repository,
	}, nil
}

func (service *Impl) GetFeedTypes() []entities.FeedType {
	return service.feedTypes
}

func (service *Impl) GetFeedType(id string) *entities.FeedType {
	for _, feedType := range service.feedTypes {
		if feedType.ID == id {
			return &feedType
		}
	}

	return nil
}

func (service *Impl) FindFeedTypes(name string, locale discordgo.Locale, limit int) []entities.FeedType {
	feedTypesFound := make([]entities.FeedType, 0)
	cleanedName, err := translators.NormalizeName(name)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize feed type name, returning empty feed type list")
		return feedTypesFound
	}

	for _, feedType := range service.feedTypes {
		currentCleanedName, errStr := translators.NormalizeName(translators.GetEntityLabel(feedType, locale))
		if errStr == nil && strings.Contains(currentCleanedName, cleanedName) {
			if currentCleanedName == cleanedName {
				return []entities.FeedType{feedType}
			}

			feedTypesFound = append(feedTypesFound, feedType)
		}
	}

	if len(feedTypesFound) > limit {
		return feedTypesFound[:limit]
	}

	return feedTypesFound
}
