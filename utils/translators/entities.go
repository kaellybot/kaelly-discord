package translators

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/rs/zerolog/log"
)

func GetEntityLabel(entity entities.LabelledEntity, locale discordgo.Locale) string {
	labels := entity.GetLabels()

	label, found := labels[i18n.MapDiscordLocale(locale)]
	if found {
		return label
	}

	log.Warn().
		Str(constants.LogEntity, entity.GetID()).
		Str(constants.LogLocale, string(locale)).
		Msgf("Entity i18n value is empty, returning value based on default locale")

	defaultLabel, found := labels[i18n.MapDiscordLocale(i18n.DefaultLocale)]
	if found {
		return defaultLabel
	}

	log.Warn().
		Str(constants.LogEntity, entity.GetID()).
		Str(constants.LogLocale, string(i18n.DefaultLocale)).
		Msgf("Entity i18n default value is empty, returning id")
	return entity.GetID()
}

func GetCitiesLabels(cities []entities.City, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, city := range cities {
		labels = append(labels, GetEntityLabel(city, locale))
	}

	return labels
}

func GetOrdersLabels(orders []entities.Order, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, order := range orders {
		labels = append(labels, GetEntityLabel(order, locale))
	}

	return labels
}

func GetFeedTypesLabels(feedTypes []entities.FeedType, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, feedType := range feedTypes {
		labels = append(labels, GetEntityLabel(feedType, locale))
	}

	return labels
}

func GetJobsLabels(jobs []entities.Job, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, job := range jobs {
		labels = append(labels, GetEntityLabel(job, locale))
	}

	return labels
}

func GetServersLabels(servers []entities.Server, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, server := range servers {
		labels = append(labels, GetEntityLabel(server, locale))
	}

	return labels
}

func GetTwittersLabels(twitterAccounts []entities.TwitterAccount, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, twitterAccount := range twitterAccounts {
		labels = append(labels, GetEntityLabel(twitterAccount, locale))
	}

	return labels
}
