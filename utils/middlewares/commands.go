package middlewares

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
)

type NextFunc func()
type MiddlewareCommand func(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale, next NextFunc)

func Use(chainedFunctions ...MiddlewareCommand) models.DiscordHandler {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, lg discordgo.Locale) {
		wrapped := func() {}
		for i := len(chainedFunctions) - 1; i >= 0; i-- {
			index := i
			currentNext := wrapped
			wrapped = func() {
				chainedFunctions[index](session, interaction, lg, currentNext)
			}
		}
		wrapped()
	}
}
