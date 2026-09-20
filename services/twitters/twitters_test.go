package twitters

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/twitters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTwitterRepo struct {
	repository.Repository
	accounts []entities.TwitterAccount
	err      error
}

func (repo *fakeTwitterRepo) GetTwitterAccounts() ([]entities.TwitterAccount, error) {
	return repo.accounts, repo.err
}

func account(id, name string) entities.TwitterAccount {
	return entities.TwitterAccount{ID: id, Name: name, NewsChannelID: id + "-chan"}
}

func newService(t *testing.T, accounts ...entities.TwitterAccount) *Impl {
	t.Helper()
	service, err := New(&fakeTwitterRepo{accounts: accounts})
	require.NoError(t, err)
	return service
}

func ids(accounts []entities.TwitterAccount) []string {
	out := make([]string, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, a.ID)
	}
	return out
}

func TestNew_PropagatesRepositoryError(t *testing.T) {
	boom := errors.New("db down")
	service, err := New(&fakeTwitterRepo{err: boom})
	require.ErrorIs(t, err, boom)
	assert.Nil(t, service)
}

func TestGetTwitterAccounts_KeepsRepositoryOrder(t *testing.T) {
	service := newService(t, account("2", "DOFUS_EN"), account("1", "DOFUSfr"))
	assert.Equal(t, []string{"2", "1"}, ids(service.GetTwitterAccounts()))
}

func TestGetTwitterAccount_ReturnsCopyOrNil(t *testing.T) {
	service := newService(t, account("1", "DOFUSfr"))

	got := service.GetTwitterAccount("1")
	require.NotNil(t, got)
	assert.Equal(t, "1-chan", got.NewsChannelID)

	got.NewsChannelID = "mutated"
	assert.Equal(t, "1-chan", service.GetTwitterAccount("1").NewsChannelID)

	assert.Nil(t, service.GetTwitterAccount("unknown"))
}

func TestFindTwitterAccounts_MatchesAccountNameInEverySupportedLocale(t *testing.T) {
	service := newService(t, account("fr", "DOFUSfr"), account("es", "ES_DOFUS"))

	for _, locale := range []discordgo.Locale{
		discordgo.French, discordgo.EnglishGB, discordgo.EnglishUS,
		discordgo.SpanishES, discordgo.German, discordgo.PortugueseBR, discordgo.Japanese,
	} {
		t.Run(string(locale), func(t *testing.T) {
			assert.Equal(t, []string{"fr"}, ids(service.FindTwitterAccounts("dofusFR", locale, 10)),
				"the account name is the label whatever the locale")
		})
	}
}

func TestFindTwitterAccounts_ExactMatchLimitAndMisses(t *testing.T) {
	service := newService(t,
		account("fr", "DOFUSfr"),
		account("main", "DOFUS"),
		account("es", "ES_DOFUS"),
		account("touch", "DofusTouch"),
	)

	assert.Equal(t, []string{"main"}, ids(service.FindTwitterAccounts("dofus", discordgo.French, 10)),
		"exact match discards partials")
	assert.Equal(t, []string{"fr", "main"}, ids(service.FindTwitterAccounts("dofu", discordgo.French, 2)))
	assert.Empty(t, service.FindTwitterAccounts("dofu", discordgo.French, 0))
	assert.Empty(t, service.FindTwitterAccounts("dofu", discordgo.French, -1), "a negative limit yields nothing")
	assert.Equal(t, []string{"fr", "main", "es", "touch"},
		ids(service.FindTwitterAccounts("DÔFU", discordgo.French, 10)), "accents in the query are ignored")

	none := service.FindTwitterAccounts("krosmoz", discordgo.French, 10)
	assert.NotNil(t, none)
	assert.Empty(t, none)
}
