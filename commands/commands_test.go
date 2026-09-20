package commands

import (
	"regexp"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func event(interactionType discordgo.InteractionType, data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{Type: interactionType, Data: data}}
}

func slash(subCommands ...string) *discordgo.InteractionCreate {
	options := make([]*discordgo.ApplicationCommandInteractionDataOption, 0, len(subCommands))
	for _, name := range subCommands {
		options = append(options, &discordgo.ApplicationCommandInteractionDataOption{
			Name: name, Type: discordgo.ApplicationCommandOptionSubCommand,
		})
	}
	return event(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: "config", Options: options})
}

func button(customID string) *discordgo.InteractionCreate {
	return event(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: customID})
}

// spy returns a handler recording its name, and asserts it got the expected arguments.
func spy(t *testing.T, name string, calls *[]string, s *discordgo.Session) DiscordHandler {
	t.Helper()
	return func(gotSession *discordgo.Session, _ *discordgo.InteractionCreate) error {
		assert.Same(t, s, gotSession)
		*calls = append(*calls, name)
		return nil
	}
}

func TestIsApplicationCommand(t *testing.T) {
	cases := map[discordgo.InteractionType]bool{
		discordgo.InteractionPing:                           false,
		discordgo.InteractionApplicationCommand:             true,
		discordgo.InteractionApplicationCommandAutocomplete: true,
		discordgo.InteractionMessageComponent:               false,
		discordgo.InteractionModalSubmit:                    false,
	}
	for interactionType, expected := range cases {
		assert.Equal(t, expected, IsApplicationCommand(event(interactionType, nil)), interactionType.String())
	}
}

func TestIsMessageCommand(t *testing.T) {
	cases := map[discordgo.InteractionType]bool{
		discordgo.InteractionPing:                           false,
		discordgo.InteractionApplicationCommand:             false,
		discordgo.InteractionApplicationCommandAutocomplete: false,
		discordgo.InteractionMessageComponent:               true,
		discordgo.InteractionModalSubmit:                    false,
	}
	for interactionType, expected := range cases {
		assert.Equal(t, expected, IsMessageCommand(event(interactionType, nil)), interactionType.String())
	}
}

func TestCallHandler_DispatchesOnInteractionType(t *testing.T) {
	session := &discordgo.Session{}
	var calls []string
	handlers := make(DiscordHandlers)
	handlers[discordgo.InteractionApplicationCommand] = spy(t, "command", &calls, session)
	handlers[discordgo.InteractionApplicationCommandAutocomplete] = spy(t, "autocomplete", &calls, session)
	command := &AbstractCommand{}

	assert.NoError(t, command.CallHandler(session,
		event(discordgo.InteractionApplicationCommandAutocomplete, nil), handlers))
	assert.NoError(t, command.CallHandler(session,
		event(discordgo.InteractionApplicationCommand, nil), handlers))

	assert.Equal(t, []string{"autocomplete", "command"}, calls)
}

func TestCallHandler_UnhandledTypeIsIgnored(t *testing.T) {
	var calls []string
	handlers := make(DiscordHandlers)
	handlers[discordgo.InteractionApplicationCommand] = spy(t, "command", &calls, nil)

	assert.NoError(t, (&AbstractCommand{}).CallHandler(nil,
		event(discordgo.InteractionMessageComponent, nil), handlers))
	assert.NoError(t, (&AbstractCommand{}).CallHandler(nil,
		event(discordgo.InteractionApplicationCommand, nil), nil))
	assert.Empty(t, calls, "an interaction type nobody registered is ignored, not reported")
}

func TestHandleSubCommands_DispatchesOnSubCommandName(t *testing.T) {
	session := &discordgo.Session{}
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{
		"get": spy(t, "get", &calls, session),
		"set": spy(t, "set", &calls, session),
	})

	assert.NoError(t, handler(session, slash("set")))
	assert.NoError(t, handler(session, slash("get")))

	assert.Equal(t, []string{"set", "get"}, calls)
}

func TestHandleSubCommands_AutocompleteIsDispatchedToo(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{"get": spy(t, "get", &calls, nil)})

	autocomplete := event(discordgo.InteractionApplicationCommandAutocomplete,
		discordgo.ApplicationCommandInteractionData{Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{Name: "get", Type: discordgo.ApplicationCommandOptionSubCommand},
		}})

	assert.NoError(t, handler(nil, autocomplete))

	assert.Equal(t, []string{"get"}, calls)
}

func TestHandleSubCommands_UnknownSubCommandIsReportedAsStale(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{"get": spy(t, "get", &calls, nil)})

	err := handler(nil, slash("delete"))

	require.ErrorIs(t, err, ErrNoProvidedHandler)
	assert.ErrorIs(t, err, ErrStaleInteraction)
	assert.Contains(t, err.Error(), "delete", "the sub-command is named for the logs")
	assert.Empty(t, calls)
}

func TestHandleSubCommands_NoSubCommandCallsNothing(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{"get": spy(t, "get", &calls, nil)})

	assert.NoError(t, handler(nil, slash()))

	assert.Empty(t, calls)
}

func TestHandleSubCommands_EveryTopLevelOptionIsDispatched(t *testing.T) {
	// Discord only ever sends one subcommand; this pins what happens otherwise.
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{
		"get": spy(t, "get", &calls, nil),
		"set": spy(t, "set", &calls, nil),
	})

	assert.NoError(t, handler(nil, slash("get", "set")))

	assert.Equal(t, []string{"get", "set"}, calls)
}

func TestHandleSubCommands_IgnoresNonApplicationInteractions(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{"get": spy(t, "get", &calls, nil)})

	assert.NoError(t, handler(nil, button("get")),
		"message component data must not be read as command data")
	assert.Empty(t, calls)
}

func TestHandleInteractionMessages_DispatchesOnCustomIDRegex(t *testing.T) {
	session := &discordgo.Session{}
	var calls []string
	handler := (&AbstractCommand{}).HandleInteractionMessages(InteractionMessageHandlers{
		regexp.MustCompile(`^/set/\d+$`):         spy(t, "set", &calls, session),
		regexp.MustCompile(`^/set/\d+/effects$`): spy(t, "effects", &calls, session),
	})

	assert.NoError(t, handler(session, button("/set/42/effects")))
	assert.NoError(t, handler(session, button("/set/42")))

	assert.Equal(t, []string{"effects", "set"}, calls)
}

func TestHandleInteractionMessages_OverlappingPatternsDispatchDeterministically(t *testing.T) {
	// Both patterns match the same custom ID: the winner must not depend on Go's
	// randomized map iteration order. Patterns are tried sorted by their string,
	// so the shorter one, a prefix of the other, always wins.
	var calls []string
	handlers := InteractionMessageHandlers{
		regexp.MustCompile(`^/set/\d+`):          spy(t, "prefix", &calls, nil),
		regexp.MustCompile(`^/set/\d+/effects$`): spy(t, "effects", &calls, nil),
	}

	const dispatches = 100
	for range dispatches {
		assert.NoError(t, (&AbstractCommand{}).HandleInteractionMessages(handlers)(nil,
			button("/set/42/effects")))
	}

	require.Len(t, calls, dispatches)
	for _, got := range calls {
		assert.Equal(t, "prefix", got, "the same handler must win on every dispatch")
	}
}

func TestHandleInteractionMessages_UnmatchedCustomIDIsReportedAsStale(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleInteractionMessages(InteractionMessageHandlers{
		regexp.MustCompile(`^/set/\d+$`): spy(t, "set", &calls, nil),
	})

	for _, customID := range []string{"/item/42", "/set/abc"} {
		err := handler(nil, button(customID))
		require.ErrorIs(t, err, ErrNoProvidedHandler)
		assert.ErrorIs(t, err, ErrStaleInteraction, "a button from an older version is not a crash")
	}
	assert.Empty(t, calls)
}

func TestHandleInteractionMessages_IgnoresApplicationCommands(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleInteractionMessages(InteractionMessageHandlers{
		regexp.MustCompile(`.*`): spy(t, "any", &calls, nil),
	})

	assert.NoError(t, handler(nil, slash("get")))
	assert.Empty(t, calls)
}
