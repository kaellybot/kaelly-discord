package translators

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/rs/zerolog/log"
)

func GetEntityLabel(entity entities.LabelledEntity, locale discordgo.Locale) string {
	labels := entity.GetLabels()

	label, found := labels[constants.MapDiscordLocale(locale)]
	if found {
		return label
	}

	log.Warn().
		Str(constants.LogEntity, entity.GetID()).
		Str(constants.LogLocale, string(locale)).
		Msgf("Entity i18n value is empty, returning value based on default locale")

	defaultLabel, found := labels[constants.MapDiscordLocale(constants.DefaultLocale)]
	if found {
		return defaultLabel
	}

	log.Warn().
		Str(constants.LogEntity, entity.GetID()).
		Str(constants.LogLocale, string(constants.DefaultLocale)).
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

func GetDimensionsLabels(dimensions []entities.Dimension, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, dimension := range dimensions {
		labels = append(labels, GetEntityLabel(dimension, locale))
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

func GetFeedTypesLabels(dimensions []entities.FeedType, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, dimension := range dimensions {
		labels = append(labels, GetEntityLabel(dimension, locale))
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

func GetStreamersLabels(streamers []entities.Streamer, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, streamer := range streamers {
		labels = append(labels, GetEntityLabel(streamer, locale))
	}

	return labels
}

func GetVideastsLabels(videasts []entities.Videast, locale discordgo.Locale) []string {
	labels := make([]string, 0)
	for _, videast := range videasts {
		labels = append(labels, GetEntityLabel(videast, locale))
	}

	return labels
}
