package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnswerError_OnlyAnswersWhatWasDeferred(t *testing.T) {
	// AnswerError edits a deferred answer; an autocomplete has none, and
	// Discord rejects an edit for it.
	cases := map[discordgo.InteractionType]bool{
		discordgo.InteractionApplicationCommand:             true,
		discordgo.InteractionMessageComponent:               true,
		discordgo.InteractionApplicationCommandAutocomplete: false,
		discordgo.InteractionModalSubmit:                    false,
		discordgo.InteractionPing:                           false,
	}
	for interactionType, answered := range cases {
		t.Run(interactionType.String(), func(t *testing.T) {
			s, fake := testutils.NewDiscordSession(t)
			event := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				ID: "i1", AppID: "app", Token: "tok", Type: interactionType,
				Locale: discordgo.French, Data: dataFor(interactionType),
			}}

			AnswerError(s, event, commands.ErrInvalidInteraction)

			assert.Equal(t, answered, len(fake.Requests()) == 1)
		})
	}
}

func TestAnswerError_CategoryPicksTheMessage(t *testing.T) {
	cases := map[string]struct {
		err error
		key string
	}{
		"stale":       {commands.ErrInvalidInteraction, "errors.unknown_interaction"},
		"unavailable": {commands.ErrUnavailable, "errors.unavailable"},
		"bug":         {commands.ErrRequestPropertyNotFound, "panic"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			s, fake := testutils.NewDiscordSession(t)
			event := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				ID: "i1", AppID: "app", Token: "tok",
				Type: discordgo.InteractionApplicationCommand, Locale: discordgo.French,
				Data: discordgo.ApplicationCommandInteractionData{Name: "item"},
			}}

			AnswerError(s, event, testCase.err)

			payload := fake.LastEdit(t)
			require.NotNil(t, payload.Content)
			assert.Contains(t, testutils.Translations(t, discordgo.French, testCase.key), *payload.Content)
		})
	}
}

// dataFor builds the payload discordgo would carry for that interaction type.
func dataFor(interactionType discordgo.InteractionType) discordgo.InteractionData {
	if interactionType == discordgo.InteractionMessageComponent {
		return discordgo.MessageComponentInteractionData{CustomID: "/item/42"}
	}

	return discordgo.ApplicationCommandInteractionData{Name: "item"}
}

func TestAnswerError_NamesTheInteractionWithoutAssumingItsPayload(t *testing.T) {
	// A payload that does not match the type must not turn error reporting
	// into a panic of its own.
	s, fake := testutils.NewDiscordSession(t)
	event := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i1", AppID: "app", Token: "tok",
		Type: discordgo.InteractionMessageComponent, Locale: discordgo.French,
		Data: discordgo.ApplicationCommandInteractionData{Name: "item"},
	}}

	assert.NotPanics(t, func() { AnswerError(s, event, commands.ErrInvalidInteraction) })
	assert.Len(t, fake.Requests(), 1)
}
