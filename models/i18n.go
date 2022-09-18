package models

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const (
	i18nFolder = "i18n"

	frenchFile  = "fr.json"
	englishFile = "en.json"

	DefaultLocale = discordgo.EnglishUS
)

var (
	TranslationFiles = map[discordgo.Locale]string{
		discordgo.French:    fmt.Sprintf("%s/%s", i18nFolder, frenchFile),
		discordgo.EnglishGB: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
		discordgo.EnglishUS: fmt.Sprintf("%s/%s", i18nFolder, englishFile),
	}
)
