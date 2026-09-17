package middlewares

import (
	"context"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

type NextFunc func(ctx context.Context)
type MiddlewareCommand func(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next NextFunc)

func Use(chainedFunctions ...MiddlewareCommand) commands.DiscordHandler {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		wrapped := func(_ context.Context) {}
		for _, chainedFunction := range slices.Backward(chainedFunctions) {
			currentNext := wrapped
			wrapped = func(ctx context.Context) {
				chainedFunction(ctx, session, interaction, currentNext)
			}
		}

		wrapped(context.Background())
	}
}
