package twitters

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/twitters"

	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
)

func New(repository repository.Repository) (*Impl, error) {
	twitterAccounts, err := repository.GetTwitterAccounts()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(twitterAccounts)).
		Msgf("Twitter Accounts loaded")

	return &Impl{
		twitterAccounts: twitterAccounts,
		repository:      repository,
	}, nil
}

func (service *Impl) GetTwitterAccounts() []entities.TwitterAccount {
	return service.twitterAccounts
}

func (service *Impl) GetTwitterAccount(id string) *entities.TwitterAccount {
	for _, twitterAccount := range service.twitterAccounts {
		if twitterAccount.ID == id {
			return &twitterAccount
		}
	}

	return nil
}

func (service *Impl) FindTwitterAccounts(name string, locale discordgo.Locale, limit int) []entities.TwitterAccount {
	twitterAccountFound := make([]entities.TwitterAccount, 0)
	cleanedName, err := translators.NormalizeName(name)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize twitter account name, returning empty twitter account name list")
		return twitterAccountFound
	}

	for _, twitterAccount := range service.twitterAccounts {
		currentCleanedName, errStr := translators.NormalizeName(translators.GetEntityLabel(twitterAccount, locale))
		if errStr == nil && strings.Contains(currentCleanedName, cleanedName) {
			if currentCleanedName == cleanedName {
				return []entities.TwitterAccount{twitterAccount}
			}

			twitterAccountFound = append(twitterAccountFound, twitterAccount)
		}
	}

	// A negative limit is not a request for more than everything: it yields nothing.
	if len(twitterAccountFound) > limit {
		return twitterAccountFound[:max(limit, 0)]
	}

	return twitterAccountFound
}
