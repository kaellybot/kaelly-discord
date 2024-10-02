package validators

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func ExpectOnlyOneElement[T any](i18nPrefix, optionValue string, collection []T,
	lg discordgo.Locale) (discordgo.WebhookEdit, bool) {
	if len(collection) == 1 {
		return discordgo.WebhookEdit{}, true
	}

	if len(collection) > 1 {
		content := i18n.Get(lg, fmt.Sprintf("%v.too_many", i18nPrefix),
			i18n.Vars{"value": optionValue, "collection": collection})
		return discordgo.WebhookEdit{
			Content: &content,
		}, false
	}

	content := i18n.Get(lg, fmt.Sprintf("%v.not_found", i18nPrefix),
		i18n.Vars{"value": optionValue})
	return discordgo.WebhookEdit{
		Content: &content,
	}, false
}

func HasWebhookPermission(s *discordgo.Session, channelID string) bool {
	permissions, err := s.State.UserChannelPermissions(s.State.User.ID, channelID)
	if err != nil {
		log.Error().Err(err).
			Str(constants.LogChannelID, channelID).
			Msg("Cannot retrieve channel permission, returning false")
		return false
	}

	return permissions&discordgo.PermissionManageWebhooks != 0
}
