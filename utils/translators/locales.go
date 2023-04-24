package translators

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func GetLocalChoices() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	for id, name := range amqp.Language_name {
		if id != int32(amqp.Language_ANY) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:              name,
				NameLocalizations: *i18n.GetLocalizations(fmt.Sprintf("locales.%s.name", name)),
				Value:             id,
			})
		}
	}

	return choices
}
