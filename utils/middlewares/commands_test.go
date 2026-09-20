package middlewares

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

type traceKey struct{}

// recorder returns a middleware that appends its name to trace, tags the
// context with it, and forwards to next unless it is a blocker.
func recorder(name string, trace *[]string, forward bool) MiddlewareCommand {
	return func(ctx context.Context, _ *discordgo.Session, _ *discordgo.InteractionCreate, next NextFunc) {
		previous, _ := ctx.Value(traceKey{}).(string)
		*trace = append(*trace, name+"<"+previous)
		if forward {
			next(context.WithValue(ctx, traceKey{}, name))
		}
	}
}

func TestUse_RunsMiddlewaresInDeclarationOrderThreadingContext(t *testing.T) {
	var trace []string
	handler := Use(recorder("a", &trace, true), recorder("b", &trace, true), recorder("c", &trace, true))

	handler(nil, &discordgo.InteractionCreate{})

	assert.Equal(t, []string{"a<", "b<a", "c<b"}, trace,
		"each middleware sees the context produced by the previous one")
}

func TestUse_MiddlewareNotCallingNextStopsTheChain(t *testing.T) {
	var trace []string
	handler := Use(recorder("check", &trace, true), recorder("guard", &trace, false),
		recorder("command", &trace, true))

	handler(nil, &discordgo.InteractionCreate{})

	assert.Equal(t, []string{"check<", "guard<check"}, trace)
}

func TestUse_PassesSessionAndInteractionToEveryMiddleware(t *testing.T) {
	session := &discordgo.Session{}
	interaction := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{ID: "i1"}}
	var seen int
	spy := func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, next NextFunc) {
		assert.Same(t, session, s)
		assert.Same(t, interaction, i)
		seen++
		next(ctx)
	}

	Use(spy, spy)(session, interaction)

	assert.Equal(t, 2, seen)
}

func TestUse_HandlerIsReusable(t *testing.T) {
	var trace []string
	handler := Use(recorder("a", &trace, true), recorder("b", &trace, true))

	handler(nil, &discordgo.InteractionCreate{})
	handler(nil, &discordgo.InteractionCreate{})

	assert.Equal(t, []string{"a<", "b<a", "a<", "b<a"}, trace, "each invocation starts from a fresh context")
}

func TestUse_LastNextIsATerminalNoOp(t *testing.T) {
	var reachedEnd bool
	last := func(ctx context.Context, _ *discordgo.Session, _ *discordgo.InteractionCreate, next NextFunc) {
		next(ctx) // must not panic nor loop back into the chain
		reachedEnd = true
	}
	var trace []string

	Use(recorder("a", &trace, true), last)(nil, nil)

	assert.True(t, reachedEnd)
	assert.Equal(t, []string{"a<"}, trace, "the chain is not re-entered")
}
