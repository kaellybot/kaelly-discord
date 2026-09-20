package mappers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/stretchr/testify/assert"
)

func TestRequestBackbone_MapsDiscordLocaleToAMQPLanguage(t *testing.T) {
	cases := []struct {
		locale   discordgo.Locale
		expected amqp.Language
	}{
		{discordgo.French, amqp.Language_FR},
		{discordgo.EnglishGB, amqp.Language_EN},
		{discordgo.EnglishUS, amqp.Language_EN},
		{discordgo.SpanishES, amqp.Language_ES},
		{discordgo.German, amqp.Language_DE},
		{discordgo.PortugueseBR, amqp.Language_PT},
		{discordgo.Japanese, amqp.Language_ANY},
		{discordgo.Unknown, amqp.Language_ANY},
	}
	for _, tc := range cases {
		t.Run(string(tc.locale), func(t *testing.T) {
			request := requestBackbone("author", amqp.RabbitMQMessage_ABOUT_REQUEST, tc.locale)
			assert.Equal(t, tc.expected, request.Language)
		})
	}
}

func TestRequestBackbone_SetsTypeGameAndAuthor(t *testing.T) {
	request := requestBackbone("author", amqp.RabbitMQMessage_JOB_SET_REQUEST, discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_JOB_SET_REQUEST, request.Type)
	assert.Equal(t, amqp.Game_DOFUS_GAME, request.Game)
	assert.Equal(t, "author", request.UserID)
}

func TestMapHelpAndAboutRequests_CarryOnlyTheBackbone(t *testing.T) {
	cases := []struct {
		expectedType amqp.RabbitMQMessage_Type
		request      *amqp.RabbitMQMessage
	}{
		{amqp.RabbitMQMessage_HELP_REQUEST, MapHelpRequest("author", discordgo.German)},
		{amqp.RabbitMQMessage_ABOUT_REQUEST, MapAboutRequest("author", discordgo.German)},
	}
	for _, tc := range cases {
		t.Run(tc.expectedType.String(), func(t *testing.T) {
			expected := &amqp.RabbitMQMessage{
				Type:     tc.expectedType,
				Language: amqp.Language_DE,
				Game:     amqp.Game_DOFUS_GAME,
				UserID:   "author",
			}
			assert.Equal(t, expected.String(), tc.request.String())
		})
	}
}
