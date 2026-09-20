package panics

import (
	"errors"
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

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i1", AppID: "app", Token: "tok", Type: interactionType,
		Locale: discordgo.French, Data: data,
	}}
}

func TestHandlePanic_NoPanic_SendsNothing(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	func() {
		defer HandlePanic(s, interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "about"}))
	}()
	assert.Empty(t, fake.Requests())
}

func TestHandlePanic_RecoversAndEditsResponseWithLocalizedMessage(t *testing.T) {
	cases := map[string]*discordgo.InteractionCreate{
		"application command": interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "about"}),
		"autocomplete": interaction(discordgo.InteractionApplicationCommandAutocomplete,
			discordgo.ApplicationCommandInteractionData{Name: "item"}),
		"message component": interaction(discordgo.InteractionMessageComponent,
			discordgo.MessageComponentInteractionData{CustomID: "/set/1"}),
		"unhandled type": interaction(discordgo.InteractionModalSubmit, nil),
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			s, fake := testutils.NewDiscordSession(t)
			require.NotPanics(t, func() {
				defer HandlePanic(s, event)
				panic("boom")
			})

			req := fake.Last(t)
			assert.Equal(t, "PATCH", req.Method)
			assert.Equal(t, discordgo.EndpointWebhookMessage("app", "tok", "@original"),
				"https://discord.com"+req.Path)
			payload := fake.LastEdit(t)
			require.NotNil(t, payload.Content)
			assert.Contains(t, testutils.Translations(t, discordgo.French, "panic"), *payload.Content,
				"the caller's locale drives the message")
			assert.Empty(t, payload.Embeds, "previous embeds are wiped")
		})
	}
}

func TestHandlePanic_DiscordFailureIsSwallowed(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 500, Body: `{"message":"x"}`})
	assert.NotPanics(t, func() {
		defer HandlePanic(s, interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "about"}))
		panic("boom")
	})
	assert.Len(t, fake.Requests(), 1)
}

func TestHandleBackgroundPanic_RecoversWithoutAnswering(t *testing.T) {
	_, fake := testutils.NewDiscordSession(t)

	assert.NotPanics(t, func() {
		defer HandleBackgroundPanic("guildCreate")
		panic(errors.New("boom"))
	})
	assert.Empty(t, fake.Requests(), "there is no interaction to answer")
}

func TestHandleBackgroundPanic_NoPanicIsANoOp(t *testing.T) {
	assert.NotPanics(t, func() {
		defer HandleBackgroundPanic("ready")
	})
}
