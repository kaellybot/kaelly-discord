package almanax

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCheck runs a check and returns the context passed to next (nil if next
// was not called) plus the recorded Discord requests.
func runCheck(t *testing.T, check func(*Command) middlewares.MiddlewareCommand,
	event *discordgo.InteractionCreate) (context.Context, *testutils.FakeDiscord) {
	t.Helper()
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)
	var got context.Context
	check(cmd)(context.Background(), s, event, func(ctx context.Context) { got = ctx })
	return got, fake
}

func checkDate(cmd *Command) middlewares.MiddlewareCommand     { return cmd.checkDate }
func checkDuration(cmd *Command) middlewares.MiddlewareCommand { return cmd.checkDuration }
func checkQuery(cmd *Command) middlewares.MiddlewareCommand    { return cmd.checkQuery }

func assertEditContent(t *testing.T, fake *testutils.FakeDiscord, expected string) {
	t.Helper()
	requests := fake.Requests()
	require.Len(t, requests, 1)
	assert.Equal(t, "PATCH", requests[0].Method)
	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	assert.Equal(t, expected, *payload.Content)
}

func TestCheckDate_DefaultsToNowWhenOptionMissingOrBlank(t *testing.T) {
	for name, event := range map[string]*discordgo.InteractionCreate{
		"missing": slash("day"),
		"blank":   slash("day", stringOption("date", "   ")),
	} {
		t.Run(name, func(t *testing.T) {
			before := time.Now()
			ctx, fake := runCheck(t, checkDate, event)

			require.NotNil(t, ctx, "next must be called")
			date, ok := ctx.Value(constants.ContextKeyDate).(*time.Time)
			require.True(t, ok)
			assert.WithinRange(t, *date, before, time.Now())
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestCheckDate_ParsesSupportedFormatsAsUTCDay(t *testing.T) {
	for _, input := range []string{"18/09/2024", "18-09-2024", "2024-09-18"} {
		t.Run(input, func(t *testing.T) {
			ctx, fake := runCheck(t, checkDate, slash("day", stringOption("date", input)))

			require.NotNil(t, ctx)
			date, ok := ctx.Value(constants.ContextKeyDate).(*time.Time)
			require.True(t, ok)
			assert.Equal(t, time.Date(2024, 9, 18, 0, 0, 0, 0, time.UTC), *date)
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestCheckDate_FirstAlmanaxDayIsAccepted(t *testing.T) {
	ctx, _ := runCheck(t, checkDate, slash("day", stringOption("date", "18/09/2012")))
	require.NotNil(t, ctx)
}

func TestCheckDate_UnparsableDateAnswersFormatError(t *testing.T) {
	ctx, fake := runCheck(t, checkDate, slash("day", stringOption("date", "tomorrow")))

	assert.Nil(t, ctx, "next must not be called")
	assertEditContent(t, fake, "La date doit respecter le format suivant : JJ/MM/AAAA :)")
}

func TestCheckDate_DateBeforeFirstAlmanaxAnswersOutOfBounds(t *testing.T) {
	ctx, fake := runCheck(t, checkDate, slash("day", stringOption("date", "17/09/2012")))

	assert.Nil(t, ctx)
	assertEditContent(t, fake, testutils.Translations(t, discordgo.French, "checks.date.outOfBounds")[0])
}

func TestCheckDuration(t *testing.T) {
	cases := map[string]struct {
		event    *discordgo.InteractionCreate
		expected int64
	}{
		"missing uses default": {slash("resources"), 7},
		"minimum":              {slash("resources", durationOption(1)), 1},
		"maximum":              {slash("resources", durationOption(30)), 30},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, fake := runCheck(t, checkDuration, tc.event)

			require.NotNil(t, ctx)
			assert.Equal(t, tc.expected, ctx.Value(constants.ContextKeyDuration))
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestCheckDuration_OutOfBoundsAnswersConstraints(t *testing.T) {
	for _, value := range []int64{0, 31} {
		ctx, fake := runCheck(t, checkDuration, slash("resources", durationOption(value)))

		assert.Nil(t, ctx)
		assertEditContent(t, fake, "La durée doit être comprise entre 1 et 30 jours :)")
	}
}

func TestCheckQuery_ForwardsRawQuery(t *testing.T) {
	ctx, fake := runCheck(t, checkQuery, slash("effects", stringOption("effect", " Pêche ")))

	require.NotNil(t, ctx)
	assert.Equal(t, " Pêche ", ctx.Value(constants.ContextKeyQuery), "query is not trimmed")
	assert.Empty(t, fake.Requests())
}

func TestCheckQuery_MissingOrBlankAnswersConstraints(t *testing.T) {
	for name, event := range map[string]*discordgo.InteractionCreate{
		"missing": slash("effects"),
		"blank":   slash("effects", stringOption("effect", "  ")),
	} {
		t.Run(name, func(t *testing.T) {
			ctx, fake := runCheck(t, checkQuery, event)

			assert.Nil(t, ctx)
			assertEditContent(t, fake, "Le champs ne doit pas être vide ! :grin:")
		})
	}
}

func TestChecks_DiscordFailureIsSwallowed(t *testing.T) {
	events := map[string]struct {
		check func(*Command) middlewares.MiddlewareCommand
		event *discordgo.InteractionCreate
	}{
		"date format": {checkDate, slash("day", stringOption("date", "nope"))},
		"date bounds": {checkDate, slash("day", stringOption("date", "01/01/2000"))},
		"duration":    {checkDuration, slash("resources", durationOption(0))},
		"query":       {checkQuery, slash("effects")},
	}
	for name, tc := range events {
		t.Run(name, func(t *testing.T) {
			cmd := newCommand(t, &fakeRequestManager{})
			s, fake := testutils.NewDiscordSession(t)
			fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 500, Body: `{}`})

			assert.NotPanics(t, func() {
				tc.check(cmd)(context.Background(), s, tc.event,
					func(context.Context) { t.Fatal("next must not be called") })
			})
			assert.Len(t, fake.Requests(), 1)
		})
	}
}
