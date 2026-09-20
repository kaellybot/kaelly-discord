package middlewares

import (
	"context"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

// NextFunc hands over to the rest of the chain. Not calling it ends the chain,
// which is what a check does once it has answered the caller itself.
type NextFunc func(ctx context.Context) error

type MiddlewareCommand func(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next NextFunc) error

func Use(chainedFunctions ...MiddlewareCommand) commands.DiscordHandler {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
		wrapped := func(_ context.Context) error { return nil }
		for _, chainedFunction := range slices.Backward(chainedFunctions) {
			currentNext := wrapped
			wrapped = func(ctx context.Context) error {
				return chainedFunction(ctx, session, interaction, currentNext)
			}
		}

		return wrapped(context.Background())
	}
}
