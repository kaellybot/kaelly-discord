package discord

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCommand matches interactions whose command name (or component custom ID
// prefix) equals its name, and records what it was asked.
type fakeCommand struct {
	commands.DiscordCommand
	name     string
	panicMsg any

	mu      sync.Mutex
	matched int
	handled []*discordgo.InteractionCreate
}

func (c *fakeCommand) Matches(i *discordgo.InteractionCreate) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.matched++
	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		return i.ApplicationCommandData().Name == c.name
	case discordgo.InteractionMessageComponent:
		return strings.HasPrefix(i.MessageComponentData().CustomID, "/"+c.name)
	case discordgo.InteractionPing, discordgo.InteractionModalSubmit:
		return false
	}
	return false
}

func (c *fakeCommand) Handle(_ *discordgo.Session, i *discordgo.InteractionCreate) {
	c.mu.Lock()
	c.handled = append(c.handled, i)
	c.mu.Unlock()
	if c.panicMsg != nil {
		panic(c.panicMsg)
	}
}

func (c *fakeCommand) handledCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.handled)
}

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i1", AppID: "app", Token: "tok", Type: interactionType,
		Locale: discordgo.French, Data: data,
	}}
}

func slash(name string) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: name})
}

const callbackPath = "/api/v9/interactions/i1/tok/callback"

func callbackType(t *testing.T, req testutils.DiscordRequest) discordgo.InteractionResponseType {
	t.Helper()
	require.Equal(t, "POST", req.Method)
	require.Equal(t, callbackPath, req.Path)
	var response discordgo.InteractionResponse
	req.Decode(t, &response)
	return response.Type
}

func TestInteractionCreate_DeferralDependsOnInteractionType(t *testing.T) {
	cases := map[string]struct {
		event    *discordgo.InteractionCreate
		deferred discordgo.InteractionResponseType // 0 means no deferral
	}{
		"slash command shows a thinking state": {
			event:    slash("about"),
			deferred: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
		"component silently acknowledges the update": {
			event: interaction(discordgo.InteractionMessageComponent,
				discordgo.MessageComponentInteractionData{CustomID: "/about/1"}),
			deferred: discordgo.InteractionResponseDeferredMessageUpdate,
		},
		"autocomplete is not deferred": {
			event: interaction(discordgo.InteractionApplicationCommandAutocomplete,
				discordgo.ApplicationCommandInteractionData{Name: "about"}),
		},
		"modal submit is not deferred": {
			event: interaction(discordgo.InteractionModalSubmit, discordgo.ModalSubmitInteractionData{}),
		},
		"ping is not deferred": {
			event: interaction(discordgo.InteractionPing, nil),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			about := &fakeCommand{name: "about"}
			service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, about)

			service.interactionCreate(service.session, tc.event)

			if tc.deferred == 0 {
				assert.Empty(t, fake.Requests())
			} else {
				requests := fake.Requests()
				require.Len(t, requests, 1)
				assert.Equal(t, tc.deferred, callbackType(t, requests[0]))
			}
		})
	}
}

func TestInteractionCreate_DispatchesToFirstMatchingCommandOnly(t *testing.T) {
	about := &fakeCommand{name: "about"}
	first := &fakeCommand{name: "item"}
	shadowed := &fakeCommand{name: "item"}
	service, _ := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, about, first, shadowed)
	event := slash("item")

	service.interactionCreate(service.session, event)

	assert.Zero(t, about.handledCount())
	require.Equal(t, 1, first.handledCount())
	assert.Same(t, event, first.handled[0])
	assert.Zero(t, shadowed.handledCount(), "the first match wins")
	assert.Zero(t, shadowed.matched, "matching stops at the first hit")
}

func TestInteractionCreate_ComponentsAreRoutedByCommand(t *testing.T) {
	about := &fakeCommand{name: "about"}
	set := &fakeCommand{name: "set"}
	service, _ := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, about, set)

	service.interactionCreate(service.session, interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: "/set/42"}))

	assert.Zero(t, about.handledCount())
	assert.Equal(t, 1, set.handledCount())
}

func TestInteractionCreate_AutocompleteIsDispatchedWithoutDeferral(t *testing.T) {
	item := &fakeCommand{name: "item"}
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, item)

	service.interactionCreate(service.session, interaction(discordgo.InteractionApplicationCommandAutocomplete,
		discordgo.ApplicationCommandInteractionData{Name: "item"}))

	assert.Equal(t, 1, item.handledCount())
	assert.Empty(t, fake.Requests(), "autocomplete answers itself")
}

func TestInteractionCreate_UnmatchedInteractionIsAnsweredWithAnError(t *testing.T) {
	cases := map[string][]commands.DiscordCommand{
		"no command registered": nil,
		"no command matches":    {&fakeCommand{name: "about"}, &fakeCommand{name: "item"}},
	}
	for name, cmds := range cases {
		t.Run(name, func(t *testing.T) {
			service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, cmds...)

			require.NotPanics(t, func() {
				service.interactionCreate(service.session, slash("unknown"))
			})

			for _, cmd := range cmds {
				fc, ok := cmd.(*fakeCommand)
				require.True(t, ok)
				assert.Equal(t, 1, fc.matched, "every command was asked")
				assert.Zero(t, fc.handledCount())
			}
			requests := fake.Requests()
			require.Len(t, requests, 2, "deferral, then the error message: no eternal thinking state")
			assert.Equal(t, discordgo.InteractionResponseDeferredChannelMessageWithSource,
				callbackType(t, requests[0]))
			assertUnknownInteractionMessage(t, fake)
		})
	}
}

func TestInteractionCreate_UnmatchedComponentIsAnsweredWithAnError(t *testing.T) {
	about := &fakeCommand{name: "about"}
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, about)

	service.interactionCreate(service.session, interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: "/unknown/1"}))

	assert.Zero(t, about.handledCount())
	requests := fake.Requests()
	require.Len(t, requests, 2)
	assert.Equal(t, discordgo.InteractionResponseDeferredMessageUpdate, callbackType(t, requests[0]))
	assertUnknownInteractionMessage(t, fake)
}

func TestInteractionCreate_UnmatchedInteractionAnswerFailureIsIgnored(t *testing.T) {
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{})
	fake.Respond(testutils.DiscordResponse{Method: "PATCH", PathContains: "/messages/@original",
		Status: 404, Body: `{"message":"Unknown Webhook","code":10015}`})

	require.NotPanics(t, func() {
		service.interactionCreate(service.session, slash("unknown"))
	})

	assert.Len(t, fake.Requests(), 2, "the failed answer is only logged")
}

func TestInteractionCreate_UnmatchedUndeferredInteractionsStaySilent(t *testing.T) {
	cases := map[string]*discordgo.InteractionCreate{
		"autocomplete": interaction(discordgo.InteractionApplicationCommandAutocomplete,
			discordgo.ApplicationCommandInteractionData{Name: "unknown"}),
		"modal submit": interaction(discordgo.InteractionModalSubmit, discordgo.ModalSubmitInteractionData{}),
		"ping":         interaction(discordgo.InteractionPing, nil),
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			about := &fakeCommand{name: "about"}
			service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, about)

			service.interactionCreate(service.session, event)

			assert.Zero(t, about.handledCount())
			assert.Empty(t, fake.Requests(), "nothing was deferred, so there is nothing to edit")
		})
	}
}

// assertEditedWith checks the deferred answer was edited with the given key's
// text, in the caller's locale.
func assertEditedWith(t *testing.T, fake *testutils.FakeDiscord, key string) {
	t.Helper()
	req := fake.Last(t)
	assert.Equal(t, "PATCH", req.Method)
	assert.Equal(t, "/api/v9/webhooks/app/tok/messages/@original", req.Path)
	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	assert.Contains(t, testutils.Translations(t, discordgo.French, key), *payload.Content)
}

func assertPanicMessage(t *testing.T, fake *testutils.FakeDiscord) {
	t.Helper()
	assertEditedWith(t, fake, "panic")
}

// assertUnknownInteractionMessage locks in that an interaction no command
// claimed gets its own message, not the crash one.
func assertUnknownInteractionMessage(t *testing.T, fake *testutils.FakeDiscord) {
	t.Helper()
	assertEditedWith(t, fake, "errors.unknown_interaction")
	assert.NotContains(t, testutils.Translations(t, discordgo.French, "panic"),
		*fake.LastEdit(t).Content, "an unhandled interaction is not a crash")
}

func TestInteractionCreate_CommandPanicIsRecoveredWithLocalizedMessage(t *testing.T) {
	cases := map[string]any{
		"string panic": "boom",
		"error panic":  errors.New("boom"),
	}
	for name, panicMsg := range cases {
		t.Run(name, func(t *testing.T) {
			about := &fakeCommand{name: "about", panicMsg: panicMsg}
			service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, about)

			require.NotPanics(t, func() {
				service.interactionCreate(service.session, slash("about"))
			})

			assert.Equal(t, 1, about.handledCount())
			requests := fake.Requests()
			require.Len(t, requests, 2, "deferral, then the panic message")
			assert.Equal(t, discordgo.InteractionResponseDeferredChannelMessageWithSource,
				callbackType(t, requests[0]))
			assertPanicMessage(t, fake)
		})
	}
}

func TestInteractionCreate_DeferralFailureAbortsDispatch(t *testing.T) {
	about := &fakeCommand{name: "about"}
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{}, about)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/callback",
		Status: 404, Body: `{"message":"Unknown interaction","code":10062}`})

	require.NotPanics(t, func() {
		service.interactionCreate(service.session, slash("about"))
	})

	assert.Zero(t, about.matched, "no command is even considered")
	assert.Zero(t, about.handledCount())
	requests := fake.Requests()
	require.Len(t, requests, 2, "failed deferral, then the panic message attempt")
	assertPanicMessage(t, fake)
}
