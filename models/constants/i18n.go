package constants

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/de"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/en_US"
	"github.com/go-playground/locales/es"
	"github.com/go-playground/locales/fr"
	"github.com/go-playground/locales/pt"
	amqp "github.com/kaellybot/kaelly-amqp"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type Language struct {
	Locale          discordgo.Locale
	Tag             language.Tag
	DateTranslator  locales.Translator
	AMQPLocale      amqp.Language
	Collator        *collate.Collator
	TranslationFile string
}

const (
	i18nFolder = "i18n"

	frenchFile     = "fr.json"
	englishFile    = "en.json"
	spanishFile    = "es.json"
	germanFile     = "de.json"
	portugueseFile = "pt.json"

	DefaultLocale = discordgo.EnglishGB
)

func GetLanguages() []Language {
	return []Language{
		{
			Locale:          discordgo.French,
			Tag:             language.French,
			DateTranslator:  fr.New(),
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, frenchFile),
			Collator:        collate.New(language.French),
			AMQPLocale:      amqp.Language_FR,
		},
		{
			Locale:          discordgo.EnglishGB,
			Tag:             language.English,
			DateTranslator:  en.New(),
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			Collator:        collate.New(language.English),
			AMQPLocale:      amqp.Language_EN,
		},
		{
			Locale:          discordgo.EnglishUS,
			Tag:             language.English,
			DateTranslator:  en_US.New(),
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
			Collator:        collate.New(language.English),
			AMQPLocale:      amqp.Language_EN,
		},
		{
			Locale:          discordgo.SpanishES,
			Tag:             language.Spanish,
			DateTranslator:  es.New(),
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, spanishFile),
			Collator:        collate.New(language.Spanish),
			AMQPLocale:      amqp.Language_ES,
		},
		{
			Locale:          discordgo.German,
			Tag:             language.German,
			DateTranslator:  de.New(),
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, germanFile),
			Collator:        collate.New(language.German),
			AMQPLocale:      amqp.Language_DE,
		},
		{
			Locale:          discordgo.PortugueseBR,
			Tag:             language.Portuguese,
			DateTranslator:  pt.New(),
			TranslationFile: fmt.Sprintf("%s/%s", i18nFolder, portugueseFile),
			Collator:        collate.New(language.Portuguese),
			AMQPLocale:      amqp.Language_PT,
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

func MapTag(locale discordgo.Locale) language.Tag {
	if locale == DefaultLocale {
		return language.English
	}

	for _, language := range GetLanguages() {
		if language.Locale == locale {
			return language.Tag
		}
	}

	return MapTag(DefaultLocale)
}

func MapCollator(locale discordgo.Locale) *collate.Collator {
	for _, language := range GetLanguages() {
		if language.Locale == locale {
			return language.Collator
		}
	}

	return getDefaultCollator()
}

func MapDateTranslator(lg discordgo.Locale) locales.Translator {
	for _, language := range GetLanguages() {
		if language.Locale == lg {
			return language.DateTranslator
		}
	}

	return getDefaultDateTranslator()
}

func MapAMQPLocale(locale amqp.Language) discordgo.Locale {
	for _, language := range GetLanguages() {
		if language.AMQPLocale == locale {
			return language.Locale
		}
	}

	return DefaultLocale
}

func getDefaultDateTranslator() locales.Translator {
	return en.New()
}

func getDefaultCollator() *collate.Collator {
	return collate.New(language.English)
}
