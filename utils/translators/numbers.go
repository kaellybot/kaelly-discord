package translators

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"golang.org/x/text/message"
)

func FormatNumber(value int64, lg discordgo.Locale) string {
	tag := constants.MapTag(lg)
	return message.NewPrinter(tag).Sprintf("%d", value)
}
