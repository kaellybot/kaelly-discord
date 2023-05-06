package middlewares

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

type NextFunc func(ctx context.Context)
type MiddlewareCommand func(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next NextFunc)

func Use(chainedFunctions ...MiddlewareCommand) commands.DiscordHandler {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, lg discordgo.Locale) {
		wrapped := func(ctx context.Context) {}
		for i := len(chainedFunctions) - 1; i >= 0; i-- {
			index := i
			currentNext := wrapped
			wrapped = func(ctx context.Context) {
				chainedFunctions[index](ctx, session, interaction, lg, currentNext)
			}
		}

		wrapped(context.Background())
	}
}
