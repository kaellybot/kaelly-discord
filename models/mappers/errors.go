package mappers

import (
	"github.com/bwmarrin/discordgo"
	di18n "github.com/kaysoro/discordgo-i18n"
)

func MapQueryMismatch(query string, lg discordgo.Locale) *discordgo.WebhookEdit {
	content := di18n.Get(lg, "errors.query_mismatch", di18n.Vars{"value": query})
	return &discordgo.WebhookEdit{
		Content: &content,
	}
}
