package mappers

import (
	"math"
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func effect(id, label string) *amqp.EncyclopediaItemAnswer_Effect {
	return &amqp.EncyclopediaItemAnswer_Effect{Id: id, Label: label}
}

func TestHighlightEffectValues_BoldsEverySignedInteger(t *testing.T) {
	cases := map[string]string{
		"10 to 15 Strength": "**10** to **15** Strength",
		"-3 AP":             "**-3** AP",
		"Invisibility":      "Invisibility",
		"":                  "",
	}
	for label, expected := range cases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, expected, highlightEffectValues(label))
		})
	}
}

func TestMapEffects_SortsByCharacteristicOrderWithUnknownLastInInputOrder(t *testing.T) {
	service := fakeCharacteristics{characs: map[string]entities.Characteristic{
		"vit": {ID: "vit", SortOrder: 2},
		"str": {ID: "str", SortOrder: 1},
	}}
	effects := []*amqp.EncyclopediaItemAnswer_Effect{
		effect("unknown-a", "A"),
		effect("vit", "100 Vitality"),
		effect("unknown-b", "B"),
		effect("str", "50 Strength"),
	}

	characs := mapEffects(effects, service, fakeEmojis{})

	assert.Equal(t, []i18nCharacteristic{
		{Label: "**50** Strength", Emoji: "<Characteristic:str>", SortOrder: 1},
		{Label: "**100** Vitality", Emoji: "<Characteristic:vit>", SortOrder: 2},
		{Label: "A", Emoji: "<Characteristic:>", SortOrder: math.MaxInt},
		{Label: "B", Emoji: "<Characteristic:>", SortOrder: math.MaxInt},
	}, characs)
}

func TestMapEffects_ReplacesRegexExpressionsWithRelativeCharacteristicEmoji(t *testing.T) {
	service := fakeCharacteristics{characs: map[string]entities.Characteristic{
		"dmg": {ID: "dmg", SortOrder: 1, Regexes: []entities.RegexCharacteristic{
			{CharacteristicID: "dmg", Expression: "(fire damage)", RelativeCharacteristicID: "fire"},
			{CharacteristicID: "dmg", Expression: "(steal)", RelativeCharacteristicID: "steal"},
		}},
	}}

	characs := mapEffects([]*amqp.EncyclopediaItemAnswer_Effect{effect("dmg", "5 (fire damage)")},
		service, fakeEmojis{})

	require.Len(t, characs, 1)
	assert.Equal(t, "**5** <Characteristic:fire>", characs[0].Label)
	assert.Equal(t, "<Characteristic:dmg>", characs[0].Emoji)
}

func TestMapEffects_EmptyInputGivesEmptySlice(t *testing.T) {
	characs := mapEffects([]*amqp.EncyclopediaItemAnswer_Effect(nil), fakeCharacteristics{}, fakeEmojis{})
	assert.NotNil(t, characs)
	assert.Empty(t, characs)
}
