package competition

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

type fakeEmojis struct {
	emojis.Service
}

func (fakeEmojis) GetMiscEmoji(emojiID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: string(emojiID)}
}

type requestCall struct {
	routingKey string
	message    *amqp.RabbitMQMessage
	callback   requests.RequestCallback
	properties []map[string]any
}

type fakeRequestManager struct {
	requests.RequestManager
	err   error
	calls []requestCall
}

func (f *fakeRequestManager) Request(_ *discordgo.Session, _ *discordgo.InteractionCreate,
	routingKey string, message *amqp.RabbitMQMessage, callback requests.RequestCallback,
	properties ...map[string]any) error {
	f.calls = append(f.calls, requestCall{routingKey, message, callback, properties})
	return f.err
}

func newCommand(t *testing.T, manager requests.RequestManager) *Command {
	t.Helper()
	viper.Set(constants.MapID, "map-discord-id")
	t.Cleanup(func() { viper.Set(constants.MapID, nil) })
	return New(fakeEmojis{}, manager)
}

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		Locale: discordgo.French, Data: data,
		Member: &discordgo.Member{User: &discordgo.User{ID: "author-id"}},
	}}
}

func mapCommand(options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: "map", Options: options})
}

func numberOption(value int64) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: "number", Type: discordgo.ApplicationCommandOptionInteger, Value: float64(value),
	}
}

func button(customID string) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: customID})
}

func mapAnswer() *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_COMPETITION_MAP_ANSWER,
		Status:   amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		CompetitionMapAnswer: &amqp.CompetitionMapAnswer{
			MapNumber:      12,
			MapNormalURL:   "https://maps/12-normal.png",
			MapTacticalURL: "https://maps/12-tactical.png",
			Source:         &amqp.Source{Name: "Krosmoz", Url: "https://source", Icon: "https://icon"},
		},
	}
}

func TestGetName_IsMap(t *testing.T) {
	assert.Equal(t, "map", newCommand(t, &fakeRequestManager{}).GetName())
}

func TestGetDescriptions_UsesConfiguredDiscordID(t *testing.T) {
	descriptions := newCommand(t, &fakeRequestManager{}).GetDescriptions(discordgo.EnglishGB)

	require.Len(t, descriptions, 1)
	assert.Equal(t, "/map", descriptions[0].Name)
	assert.Equal(t, "</map:map-discord-id>", descriptions[0].CommandID)
	assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "map.help.detailed")[0],
		descriptions[0].Description)
	assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "map.help.tutorial")[0],
		descriptions[0].TutorialURL)
}

func TestMatches(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	cases := []struct {
		name     string
		event    *discordgo.InteractionCreate
		expected bool
	}{
		{"map slash command", mapCommand(), true},
		{"other slash command", interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "almanax"}), false},
		{"normal map button", button("/map/3?type=normal"), true},
		{"tactical map button", button("/map/3?type=tactical"), true},
		{"foreign button", button("/almanax/day/1"), false},
		{"modal submit", interaction(discordgo.InteractionModalSubmit,
			discordgo.ModalSubmitInteractionData{CustomID: "/map/3?type=normal"}), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, cmd.Matches(tc.event))
		})
	}
}

func TestCheckOptionalMapNumber_ValidOrAbsentNumberReachesNext(t *testing.T) {
	cases := map[string]struct {
		event    *discordgo.InteractionCreate
		expected int64
	}{
		"absent option means random (0)": {mapCommand(), 0},
		"lower bound":                    {mapCommand(numberOption(constants.MapNumberMin)), constants.MapNumberMin},
		"upper bound":                    {mapCommand(numberOption(constants.MapNumberMax)), constants.MapNumberMax},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := newCommand(t, &fakeRequestManager{})
			s, fake := testutils.NewDiscordSession(t)
			var got any
			called := false

			cmd.checkOptionalMapNumber(context.Background(), s, tc.event, func(ctx context.Context) {
				called = true
				got = ctx.Value(constants.ContextKeyMap)
			})

			require.True(t, called)
			assert.Equal(t, tc.expected, got)
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestCheckOptionalMapNumber_OutOfBoundsAnswersLocalizedErrorAndStops(t *testing.T) {
	for _, value := range []int64{constants.MapNumberMin - 1, constants.MapNumberMax + 1} {
		cmd := newCommand(t, &fakeRequestManager{})
		s, fake := testutils.NewDiscordSession(t)

		cmd.checkOptionalMapNumber(context.Background(), s, mapCommand(numberOption(value)),
			func(context.Context) { t.Fatal("next must not be called") })

		payload := fake.LastEdit(t)
		require.NotNil(t, payload.Content)
		assert.Equal(t, "Le numéro de carte doit être compris entre 1 et 50 ! :grin:", *payload.Content)
	}
}

func TestHandle_SlashCommandRequestsMapWithCheckedNumber(t *testing.T) {
	manager := &fakeRequestManager{}
	cmd := newCommand(t, manager)
	s, fake := testutils.NewDiscordSession(t)

	cmd.Handle(s, mapCommand(numberOption(12)))

	assert.Empty(t, fake.Requests())
	require.Len(t, manager.calls, 1)
	call := manager.calls[0]
	assert.Equal(t, "requests.competitions", call.routingKey)
	assert.Equal(t, amqp.RabbitMQMessage_COMPETITION_MAP_REQUEST, call.message.Type)
	assert.Equal(t, "author-id", call.message.UserID)
	assert.Equal(t, amqp.Language_FR, call.message.Language)
	assert.Equal(t, int64(12), call.message.GetCompetitionMapRequest().GetMapNumber())
	assert.Empty(t, call.properties)

	// The slash command reply always shows the normal map.
	call.callback(context.Background(), s, mapCommand(), mapAnswer(), nil)
	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	assert.Equal(t, "https://maps/12-normal.png", payload.Embeds[0].Image.URL)
	assert.Equal(t, "Map 12", payload.Embeds[0].Title, "answer language drives the reply locale")
	assert.JSONEq(t, `[{"type":1,"components":[{"type":2,"custom_id":"/map/12?type=tactical",`+
		`"label":"Tactical mode","style":1,"disabled":false,"emoji":{"name":"tacticalMap"}}]}]`,
		componentsJSON(payload))
}

func TestHandle_SlashCommandInvalidNumberDoesNotRequest(t *testing.T) {
	manager := &fakeRequestManager{}
	cmd := newCommand(t, manager)
	s, fake := testutils.NewDiscordSession(t)

	cmd.Handle(s, mapCommand(numberOption(0)))

	assert.Empty(t, manager.calls)
	assert.Len(t, fake.Requests(), 1)
}

func TestHandle_SlashCommandRequestFailurePanics(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{err: errors.New("broker down")})
	s, _ := testutils.NewDiscordSession(t)

	assert.PanicsWithError(t, "broker down", func() { cmd.Handle(s, mapCommand()) })
}

func TestHandle_ButtonsRequestTheMapWithTheirType(t *testing.T) {
	cases := map[string]struct {
		customID    string
		mapNumber   int64
		mapType     constants.MapType
		expectedURL string
		nextButton  string
	}{
		"normal": {"/map/7?type=normal", 7, constants.MapTypeNormal, "https://maps/12-normal.png", "/map/12?type=tactical"},
		"tactical": {"/map/9?type=tactical", 9, constants.MapTypeTactical,
			"https://maps/12-tactical.png", "/map/12?type=normal"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			manager := &fakeRequestManager{}
			cmd := newCommand(t, manager)
			s, fake := testutils.NewDiscordSession(t)

			cmd.Handle(s, button(tc.customID))

			require.Len(t, manager.calls, 1)
			call := manager.calls[0]
			assert.Equal(t, "requests.competitions", call.routingKey)
			assert.Equal(t, tc.mapNumber, call.message.GetCompetitionMapRequest().GetMapNumber())
			require.Len(t, call.properties, 1)
			assert.Equal(t, map[string]any{"mapType": tc.mapType}, call.properties[0])

			call.callback(context.Background(), s, button(tc.customID), mapAnswer(), call.properties[0])
			payload := fake.LastEdit(t)
			require.Len(t, payload.Embeds, 1)
			assert.Equal(t, tc.expectedURL, payload.Embeds[0].Image.URL)
			assert.Contains(t, componentsJSON(payload), `"custom_id":"`+tc.nextButton+`"`)
		})
	}
}

func TestUpdateMap_UnknownCustomIDPanics(t *testing.T) {
	manager := &fakeRequestManager{}
	cmd := newCommand(t, manager)
	s, _ := testutils.NewDiscordSession(t)

	assert.PanicsWithValue(t, commands.ErrInvalidInteraction, func() {
		cmd.updateMap(s, button("/map/7?type=unknown"))
	})
	assert.Empty(t, manager.calls)
}

func TestUpdateMap_RequestFailurePanics(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{err: errors.New("broker down")})
	s, _ := testutils.NewDiscordSession(t)

	assert.PanicsWithError(t, "broker down", func() { cmd.Handle(s, button("/map/7?type=normal")) })
}

func TestUpdateMapReply_InvalidAnswersPanic(t *testing.T) {
	failed := mapAnswer()
	failed.Status = amqp.RabbitMQMessage_FAILED
	wrongType := mapAnswer()
	wrongType.Type = amqp.RabbitMQMessage_ABOUT_REQUEST
	missing := mapAnswer()
	missing.CompetitionMapAnswer = nil

	for name, message := range map[string]*amqp.RabbitMQMessage{
		"failed status": failed, "wrong type": wrongType, "missing payload": missing,
	} {
		t.Run(name, func(t *testing.T) {
			cmd := newCommand(t, &fakeRequestManager{})
			s, fake := testutils.NewDiscordSession(t)

			assert.PanicsWithValue(t, commands.ErrInvalidAnswerMessage, func() {
				cmd.updateMapReply(context.Background(), s, button("/map/1?type=normal"), message,
					map[string]any{"mapType": constants.MapTypeNormal})
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestUpdateMapReply_MissingOrMistypedPropertyPanics(t *testing.T) {
	for name, properties := range map[string]map[string]any{
		"missing":    {},
		"not a type": {"mapType": "normal"},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := newCommand(t, &fakeRequestManager{})
			s, fake := testutils.NewDiscordSession(t)

			assert.PanicsWithValue(t, commands.ErrRequestPropertyNotFound, func() {
				cmd.updateMapReply(context.Background(), s, button("/map/1?type=normal"), mapAnswer(), properties)
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestUpdateMapReply_DiscordFailureIsSwallowed(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 500, Body: `{}`})

	assert.NotPanics(t, func() {
		cmd.getMapReply(context.Background(), s, mapCommand(), mapAnswer(), nil)
	})
	assert.Len(t, fake.Requests(), 1)
}

func TestGetMap_MissingContextValuePanics(t *testing.T) {
	manager := &fakeRequestManager{}
	cmd := newCommand(t, manager)
	s, _ := testutils.NewDiscordSession(t)

	assert.Panics(t, func() { cmd.getMap(context.Background(), s, mapCommand(), nil) })
	assert.Empty(t, manager.calls)
}

func componentsJSON(payload testutils.WebhookEditPayload) string {
	raw, _ := json.Marshal(payload.Components)
	return string(raw)
}
