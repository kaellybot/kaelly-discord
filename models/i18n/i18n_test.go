package i18n

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestGetLanguages_LocalesAreUniqueAndDefaultIsSupported(t *testing.T) {
	seen := make(map[discordgo.Locale]bool)
	for _, lg := range GetLanguages() {
		assert.False(t, seen[lg.Locale], "locale %v declared twice", lg.Locale)
		seen[lg.Locale] = true
		assert.NotNil(t, lg.DateTranslator, lg.Locale)
		assert.NotNil(t, lg.Collator, lg.Locale)
	}
	assert.True(t, seen[DefaultLocale], "the default locale must be one of the supported languages")
	assert.Len(t, seen, 6)
}

func TestMapDiscordLocale(t *testing.T) {
	cases := []struct {
		locale   discordgo.Locale
		expected amqp.Language
	}{
		{discordgo.French, amqp.Language_FR},
		{discordgo.EnglishGB, amqp.Language_EN},
		{discordgo.EnglishUS, amqp.Language_EN},
		{discordgo.SpanishES, amqp.Language_ES},
		{discordgo.German, amqp.Language_DE},
		{discordgo.PortugueseBR, amqp.Language_PT},
		{discordgo.Japanese, amqp.Language_ANY},
		{"", amqp.Language_ANY},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.expected, MapDiscordLocale(tc.locale), "locale %q", tc.locale)
	}
}

func TestMapAMQPLocale_FirstMatchingLocaleWinsAndUnknownFallsBackToDefault(t *testing.T) {
	cases := map[amqp.Language]discordgo.Locale{
		amqp.Language_FR:  discordgo.French,
		amqp.Language_EN:  discordgo.EnglishGB, // EN is shared by GB and US: GB is declared first.
		amqp.Language_ES:  discordgo.SpanishES,
		amqp.Language_DE:  discordgo.German,
		amqp.Language_PT:  discordgo.PortugueseBR,
		amqp.Language_ANY: DefaultLocale,
	}
	for amqpLocale, expected := range cases {
		assert.Equal(t, expected, MapAMQPLocale(amqpLocale), "amqp locale %v", amqpLocale)
	}
}

func TestMapTag(t *testing.T) {
	cases := []struct {
		locale   discordgo.Locale
		expected language.Tag
	}{
		{discordgo.French, language.French},
		{discordgo.EnglishGB, language.English},
		{discordgo.EnglishUS, language.English},
		{discordgo.SpanishES, language.Spanish},
		{discordgo.German, language.German},
		{discordgo.PortugueseBR, language.Portuguese},
		{discordgo.Korean, language.English},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.expected, MapTag(tc.locale), "locale %q", tc.locale)
	}
}

func TestMapCollator_UsesLocaleRulesAndFallsBackToEnglish(t *testing.T) {
	// "ä" sorts right after "a" in German, but after "z" in Swedish; both German
	// and English collators place it before "b".
	assert.Equal(t, -1, MapCollator(discordgo.German).CompareString("ä", "b"))
	assert.Equal(t, -1, MapCollator(discordgo.Swedish).CompareString("ä", "b"),
		"unsupported locales use the English collator")
	// Accents are a secondary difference, not a byte-order one: "côte" still sorts before "cotte".
	assert.Equal(t, -1, MapCollator(discordgo.French).CompareString("côte", "cotte"))
}

func TestMapDateTranslator(t *testing.T) {
	assert.Equal(t, "fr", MapDateTranslator(discordgo.French).Locale())
	assert.Equal(t, "en_US", MapDateTranslator(discordgo.EnglishUS).Locale())
	assert.Equal(t, "en", MapDateTranslator(discordgo.EnglishGB).Locale())
	assert.Equal(t, "pt", MapDateTranslator(discordgo.PortugueseBR).Locale())
	assert.Equal(t, "en", MapDateTranslator(discordgo.Japanese).Locale(), "fallback is English")
}

func TestTranslationFiles_AreEmbeddedValidJSON(t *testing.T) {
	for _, lg := range GetLanguages() {
		raw, err := Folder.ReadFile(lg.TranslationFile)
		require.NoError(t, err, lg.TranslationFile)
		var content map[string]any
		require.NoError(t, json.Unmarshal(raw, &content), lg.TranslationFile)
		assert.NotEmpty(t, content, lg.TranslationFile)
	}
}

func TestTranslationFiles_ShareTheDefaultKeySet(t *testing.T) {
	reference := translationKeys(t, englishFile)
	for _, file := range []string{frenchFile, spanishFile, germanFile, portugueseFile} {
		t.Run(file, func(t *testing.T) {
			keys := translationKeys(t, file)
			assert.Empty(t, difference(reference, keys), "keys missing from %s", file)
			assert.Empty(t, difference(keys, reference), "keys only present in %s", file)
		})
	}
}

func translationKeys(t *testing.T, file string) map[string]bool {
	t.Helper()
	raw, err := Folder.ReadFile(file)
	require.NoError(t, err)
	var content any
	require.NoError(t, json.Unmarshal(raw, &content))
	keys := make(map[string]bool)
	collectKeys(content, nil, keys)
	return keys
}

func collectKeys(node any, path []string, keys map[string]bool) {
	object, ok := node.(map[string]any)
	if !ok {
		keys[strings.Join(path, ".")] = true
		return
	}
	for key, child := range object {
		collectKeys(child, append(append([]string(nil), path...), key), keys)
	}
}

func difference(left, right map[string]bool) []string {
	missing := make([]string, 0)
	for key := range left {
		if !right[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}
