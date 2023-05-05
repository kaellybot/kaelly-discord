package constants

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
)

type Language struct {
	Locale          discordgo.Locale
	AMQPLocale      amqp.Language
	TranslationFile string
}

const (
	i18nFolder = "i18n"

	frenchFile  = "fr.json"
	englishFile = "en.json"
	spanishFile = "es.json"

	DefaultLocale = discordgo.EnglishGB
)

var (
	Languages = []Language{
		{
			Locale:          discordgo.French,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, frenchFile),
			AMQPLocale:      amqp.Language_FR,
		},
		{
			Locale:          discordgo.EnglishGB,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			AMQPLocale:      amqp.Language_EN,
		},
		{
			Locale:          discordgo.EnglishUS,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			AMQPLocale:      amqp.Language_EN,
		},
		{
			Locale:          discordgo.SpanishES,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, spanishFile),
			AMQPLocale:      amqp.Language_ES,
		},
	}
)

func MapDiscordLocale(locale discordgo.Locale) amqp.Language {
	for _, language := range Languages {
		if language.Locale == locale {
			return language.AMQPLocale
		}
	}

	return amqp.Language_ANY
}

func MapAMQPLocale(locale amqp.Language) discordgo.Locale {
	for _, language := range Languages {
		if language.AMQPLocale == locale {
			return language.Locale
		}
	}

	return DefaultLocale
}
