package constants

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type Language struct {
	Locale          discordgo.Locale
	AMQPLocale      amqp.Language
	Collator        *collate.Collator
	TranslationFile string
}

const (
	i18nFolder = "i18n"

	frenchFile  = "fr.json"
	englishFile = "en.json"
	spanishFile = "es.json"
	germanFile  = "de.json"

	DefaultLocale = discordgo.EnglishGB
)

func GetLanguages() []Language {
	return []Language{
		{
			Locale:          discordgo.French,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, frenchFile),
			Collator:        collate.New(language.French),
			AMQPLocale:      amqp.Language_FR,
		},
		{
			Locale:          discordgo.EnglishGB,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			Collator:        collate.New(language.English),
			AMQPLocale:      amqp.Language_EN,
		},
		{
			Locale:          discordgo.EnglishUS,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			Collator:        collate.New(language.English),
			AMQPLocale:      amqp.Language_EN,
		},
		{
			Locale:          discordgo.SpanishES,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, spanishFile),
			Collator:        collate.New(language.Spanish),
			AMQPLocale:      amqp.Language_ES,
		},
		{
			Locale:          discordgo.German,
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, germanFile),
			Collator:        collate.New(language.German),
			AMQPLocale:      amqp.Language_DE,
		},
	}
}

func MapDiscordLocale(locale discordgo.Locale) amqp.Language {
	for _, language := range GetLanguages() {
		if language.Locale == locale {
			return language.AMQPLocale
		}
	}

	return amqp.Language_ANY
}

func MapCollator(locale discordgo.Locale) *collate.Collator {
	for _, language := range GetLanguages() {
		if language.Locale == locale {
			return language.Collator
		}
	}

	return getDefaultCollator()
}

func MapAMQPLocale(locale amqp.Language) discordgo.Locale {
	for _, language := range GetLanguages() {
		if language.AMQPLocale == locale {
			return language.Locale
		}
	}

	return DefaultLocale
}

func getDefaultCollator() *collate.Collator {
	return collate.New(language.English)
}
