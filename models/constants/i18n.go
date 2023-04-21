package constants

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
)

type Language struct {
	Locale          discordgo.Locale
	AmqpLocale      amqp.Language
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
			AmqpLocale:      amqp.Language_FR,
		},
		{
			Locale:          discordgo.EnglishGB,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			AmqpLocale:      amqp.Language_EN,
		},
		{
			Locale:          discordgo.EnglishUS,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			AmqpLocale:      amqp.Language_EN,
		},
		{
			Locale:          discordgo.SpanishES,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, spanishFile),
			AmqpLocale:      amqp.Language_ES,
		},
	}
)

func MapDiscordLocale(locale discordgo.Locale) amqp.Language {
	for _, language := range Languages {
		if language.Locale == locale {
			return language.AmqpLocale
		}
	}

	return amqp.Language_ANY
}

func MapAmqpLocale(locale amqp.Language) discordgo.Locale {
	for _, language := range Languages {
		if language.AmqpLocale == locale {
			return language.Locale
		}
	}

	return DefaultLocale
}
