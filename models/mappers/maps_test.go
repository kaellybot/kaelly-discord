package mappers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func competitionMap() *amqp.CompetitionMapAnswer {
	return &amqp.CompetitionMapAnswer{
		MapNumber:      17,
		MapNormalURL:   "normal.png",
		MapTacticalURL: "tactical.png",
		Source:         source(),
	}
}

func TestMapCompetitionMapRequest_CarriesMapNumber(t *testing.T) {
	request := MapCompetitionMapRequest(17, "author", discordgo.SpanishES)

	assert.Equal(t, amqp.RabbitMQMessage_COMPETITION_MAP_REQUEST, request.Type)
	assert.Equal(t, amqp.Language_ES, request.Language)
	assert.Equal(t, "author", request.UserID)
	require.NotNil(t, request.CompetitionMapRequest)
	assert.Equal(t, int64(17), request.CompetitionMapRequest.MapNumber)
}

func TestMapCompetitionMapToWebhookEdit_NormalShowsNormalImageAndOffersTactical(t *testing.T) {
	edit := MapCompetitionMapToWebhookEdit(competitionMap(), constants.MapTypeNormal, fakeEmojis{}, amqp.Language_EN)

	embed := singleEmbed(t, edit.Embeds)
	assert.Equal(t, "normal.png", embed.Image.URL)
	assert.Contains(t, embed.Title, "17")
	assertRendered(t, embed.Title)
	assert.Contains(t, testutils.Translations(t, discordgo.EnglishGB, "map.taunt"), embed.Description)
	assert.Equal(t, constants.Color, embed.Color)
	assertSourceAuthor(t, embed.Author)
	assertDefaultFooter(t, embed.Footer)

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 1)
	btns := buttons(t, rows[0])
	require.Len(t, btns, 1)
	mapNumber, ok := contract.ExtractMapTacticalCustomID(btns[0].CustomID)
	require.True(t, ok, "custom ID %q is a tactical map ID", btns[0].CustomID)
	assert.Equal(t, int64(17), mapNumber)
	assertTranslated(t, discordgo.EnglishGB, "map.button.tactical", btns[0].Label)
	assert.Equal(t, discordgo.PrimaryButton, btns[0].Style)
	assert.Equal(t, "misc:tacticalMap", btns[0].Emoji.Name)
}

func TestMapCompetitionMapToWebhookEdit_TacticalShowsTacticalImageAndOffersNormal(t *testing.T) {
	edit := MapCompetitionMapToWebhookEdit(competitionMap(), constants.MapTypeTactical, fakeEmojis{}, amqp.Language_FR)

	embed := singleEmbed(t, edit.Embeds)
	assert.Equal(t, "tactical.png", embed.Image.URL)
	assert.Contains(t, testutils.Translations(t, discordgo.French, "map.taunt"), embed.Description)

	btns := buttons(t, actionRows(t, edit.Components)[0])
	require.Len(t, btns, 1)
	mapNumber, ok := contract.ExtractMapNormalCustomID(btns[0].CustomID)
	require.True(t, ok, "custom ID %q is a normal map ID", btns[0].CustomID)
	assert.Equal(t, int64(17), mapNumber)
	assertTranslated(t, discordgo.French, "map.button.normal", btns[0].Label)
	assert.Equal(t, "misc:normalMap", btns[0].Emoji.Name)
}

func TestMapCompetitionMapToWebhookEdit_UnknownTypeHasNoImageNorButton(t *testing.T) {
	edit := MapCompetitionMapToWebhookEdit(competitionMap(), constants.MapType("other"), fakeEmojis{}, amqp.Language_ANY)

	embed := singleEmbed(t, edit.Embeds)
	assert.Empty(t, embed.Image.URL)
	assert.Contains(t, testutils.Translations(t, discordgo.EnglishGB, "map.taunt"), embed.Description,
		"unsupported AMQP language falls back to the default locale")
	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 1)
	// NOTE: an empty actions row is still emitted; Discord rejects such rows.
	assert.Empty(t, rows[0].Components)
}
