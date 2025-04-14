package translators

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"golang.org/x/text/message"
)

func FormatNumber(value int64, lg discordgo.Locale) string {
	tag := i18n.MapTag(lg)
	return message.NewPrinter(tag).Sprintf("%d", value)
}
