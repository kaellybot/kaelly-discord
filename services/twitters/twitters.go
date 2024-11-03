package twitters

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/twitters"

	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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
		transformer:     transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
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

func (service *Impl) FindTwitterAccounts(name string, locale discordgo.Locale) []entities.TwitterAccount {
	twitterAccountFound := make([]entities.TwitterAccount, 0)
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize twitter account name, returning empty twitter account name list")
		return twitterAccountFound
	}

	for _, twitterAccount := range service.twitterAccounts {
		currentCleanedName, _, errStr := transform.String(service.transformer,
			strings.ToLower(translators.GetEntityLabel(twitterAccount, locale)))
		if errStr == nil && strings.Contains(currentCleanedName, cleanedName) {
			if currentCleanedName == cleanedName {
				return []entities.TwitterAccount{twitterAccount}
			}

			twitterAccountFound = append(twitterAccountFound, twitterAccount)
		}
	}

	return twitterAccountFound
}
