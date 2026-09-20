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
	return func(gotSession *discordgo.Session, _ *discordgo.InteractionCreate) {
		assert.Same(t, s, gotSession)
		*calls = append(*calls, name)
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

	command.CallHandler(session, event(discordgo.InteractionApplicationCommandAutocomplete, nil), handlers)
	command.CallHandler(session, event(discordgo.InteractionApplicationCommand, nil), handlers)

	assert.Equal(t, []string{"autocomplete", "command"}, calls)
}

func TestCallHandler_UnhandledTypeIsIgnored(t *testing.T) {
	var calls []string
	handlers := make(DiscordHandlers)
	handlers[discordgo.InteractionApplicationCommand] = spy(t, "command", &calls, nil)

	assert.NotPanics(t, func() {
		(&AbstractCommand{}).CallHandler(nil, event(discordgo.InteractionMessageComponent, nil), handlers)
		(&AbstractCommand{}).CallHandler(nil, event(discordgo.InteractionApplicationCommand, nil), nil)
	})
	assert.Empty(t, calls)
}

func TestHandleSubCommands_DispatchesOnSubCommandName(t *testing.T) {
	session := &discordgo.Session{}
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{
		"get": spy(t, "get", &calls, session),
		"set": spy(t, "set", &calls, session),
	})

	handler(session, slash("set"))
	handler(session, slash("get"))

	assert.Equal(t, []string{"set", "get"}, calls)
}

func TestHandleSubCommands_AutocompleteIsDispatchedToo(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{"get": spy(t, "get", &calls, nil)})

	handler(nil, event(discordgo.InteractionApplicationCommandAutocomplete,
		discordgo.ApplicationCommandInteractionData{Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{Name: "get", Type: discordgo.ApplicationCommandOptionSubCommand},
		}}))

	assert.Equal(t, []string{"get"}, calls)
}

func TestHandleSubCommands_UnknownSubCommandPanics(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{"get": spy(t, "get", &calls, nil)})

	assert.PanicsWithValue(t, ErrNoProvidedHandler, func() { handler(nil, slash("delete")) })
	assert.Empty(t, calls)
}

func TestHandleSubCommands_NoSubCommandCallsNothing(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{"get": spy(t, "get", &calls, nil)})

	handler(nil, slash())

	assert.Empty(t, calls)
}

func TestHandleSubCommands_EveryTopLevelOptionIsDispatched(t *testing.T) {
	// Discord only ever sends one subcommand; this pins what happens otherwise.
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{
		"get": spy(t, "get", &calls, nil),
		"set": spy(t, "set", &calls, nil),
	})

	handler(nil, slash("get", "set"))

	assert.Equal(t, []string{"get", "set"}, calls)
}

func TestHandleSubCommands_IgnoresNonApplicationInteractions(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleSubCommands(SubCommandHandlers{"get": spy(t, "get", &calls, nil)})

	assert.NotPanics(t, func() { handler(nil, button("get")) },
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

	handler(session, button("/set/42/effects"))
	handler(session, button("/set/42"))

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
		(&AbstractCommand{}).HandleInteractionMessages(handlers)(nil, button("/set/42/effects"))
	}

	require.Len(t, calls, dispatches)
	for _, got := range calls {
		assert.Equal(t, "prefix", got, "the same handler must win on every dispatch")
	}
}

func TestHandleInteractionMessages_UnmatchedCustomIDPanics(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleInteractionMessages(InteractionMessageHandlers{
		regexp.MustCompile(`^/set/\d+$`): spy(t, "set", &calls, nil),
	})

	assert.PanicsWithValue(t, ErrNoProvidedHandler, func() { handler(nil, button("/item/42")) })
	assert.PanicsWithValue(t, ErrNoProvidedHandler, func() { handler(nil, button("/set/abc")) })
	assert.Empty(t, calls)
}

func TestHandleInteractionMessages_IgnoresApplicationCommands(t *testing.T) {
	var calls []string
	handler := (&AbstractCommand{}).HandleInteractionMessages(InteractionMessageHandlers{
		regexp.MustCompile(`.*`): spy(t, "any", &calls, nil),
	})

	assert.NotPanics(t, func() { handler(nil, slash("get")) })
	assert.Empty(t, calls)
}
