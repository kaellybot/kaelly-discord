package validators

import (
	"github.com/bwmarrin/discordgo"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func ExpectOnlyOneElement[T any](i18nPrefix, optionValue string, collection []T, lg discordgo.Locale) (discordgo.InteractionResponseData, bool) {
	if len(collection) == 1 {
		return discordgo.InteractionResponseData{}, true
	}

	if len(collection) > 1 {
		return discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: i18n.Get(lg, i18nPrefix+".too_many", i18n.Vars{"value": optionValue, "collection": collection}),
		}, false
	}

	return discordgo.InteractionResponseData{
		Flags:   discordgo.MessageFlagsEphemeral,
		Content: i18n.Get(lg, i18nPrefix+".not_found", i18n.Vars{"value": optionValue}),
	}, false
}
