package middlewares

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

type NextFunc func(ctx context.Context)
type MiddlewareCommand func(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next NextFunc)

func Use(chainedFunctions ...MiddlewareCommand) commands.DiscordHandler {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		wrapped := func(ctx context.Context) {}
		for i := len(chainedFunctions) - 1; i >= 0; i-- {
			index := i
			currentNext := wrapped
			wrapped = func(ctx context.Context) {
				chainedFunctions[index](ctx, session, interaction, currentNext)
			}
		}

		wrapped(context.Background())
	}
}
