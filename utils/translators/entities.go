package translators

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/rs/zerolog/log"
)

func GetEntityLabel(entity entities.LabelledEntity, locale discordgo.Locale) string {
	labels := entity.GetLabels()

	label, found := labels[locale]
	if found {
		return label
	}

	log.Warn().
		Str(constants.LogEntity, entity.GetId()).
		Str(constants.LogLocale, string(locale)).
		Msgf("Entity i18n value is empty, returning value based on default locale")

	defaultLabel, found := labels[locale]
	if found {
		return defaultLabel
	}

	log.Warn().
		Str(constants.LogEntity, entity.GetId()).
		Str(constants.LogLocale, string(constants.DefaultLocale)).
		Msgf("Entity i18n default value is empty, returning id")
	return entity.GetId()
}
