package validators

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

func TestExpectOnlyOneElement_SingleElementSucceedsWithEmptyResponse(t *testing.T) {
	response, ok := ExpectOnlyOneElement("checks.server", "djaul", []string{"Djaul"}, discordgo.French)

	assert.True(t, ok)
	assert.Equal(t, discordgo.WebhookEdit{}, response)
}

func TestExpectOnlyOneElement_EmptyCollectionIsNotFound(t *testing.T) {
	for name, collection := range map[string][]string{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			response, ok := ExpectOnlyOneElement("checks.server", "xyz", collection, discordgo.EnglishGB)

			assert.False(t, ok)
			require.NotNil(t, response.Content)
			assert.Equal(t, "Aie! '*xyz*' does not correspond to any servers I know! :pensive:", *response.Content)
			assert.Nil(t, response.Embeds)
		})
	}
}

func TestExpectOnlyOneElement_SeveralElementsListsThemAll(t *testing.T) {
	response, ok := ExpectOnlyOneElement("checks.server", "d", []string{"Djaul", "Dakal", "Draconiros"},
		discordgo.French)

	assert.False(t, ok)
	require.NotNil(t, response.Content)
	assert.Equal(t, "Outch! '*d*' correspond à plusieurs serveurs (Djaul, Dakal, Draconiros)! "+
		"Réessaie en sélectionnant un serveur de la liste :slight_smile:", *response.Content)
}

func TestExpectOnlyOneElement_PrefixSelectsTheTranslationFamily(t *testing.T) {
	response, ok := ExpectOnlyOneElement("checks.job", "alchi", []int{}, discordgo.EnglishGB)

	assert.False(t, ok)
	require.NotNil(t, response.Content)
	assert.Equal(t, "Aie! '*alchi*' does not correspond to any jobs I know! :pensive:", *response.Content)
}
