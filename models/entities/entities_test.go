package entities

import (
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/stretchr/testify/assert"
)

// labels builds a locale->label map from alternating (amqp.Language, string) pairs.
func labels(pairs ...any) map[amqp.Language]string {
	result := make(map[amqp.Language]string)
	for i := 0; i < len(pairs); i += 2 {
		locale, _ := pairs[i].(amqp.Language)
		label, _ := pairs[i+1].(string)
		result[locale] = label
	}
	return result
}

func TestLabelledEntities_IndexLabelsByLocale(t *testing.T) {
	expected := labels(amqp.Language_FR, "fr-label", amqp.Language_EN, "en-label")
	cases := map[string]LabelledEntity{
		"city": City{ID: "bonta", Labels: []CityLabel{
			{Locale: amqp.Language_FR, Label: "fr-label"}, {Locale: amqp.Language_EN, Label: "en-label"}}},
		"feed type": FeedType{ID: "bonta", Labels: []FeedTypeLabel{
			{Locale: amqp.Language_FR, Label: "fr-label"}, {Locale: amqp.Language_EN, Label: "en-label"}}},
		"job": Job{ID: "bonta", Labels: []JobLabel{
			{Locale: amqp.Language_FR, Label: "fr-label"}, {Locale: amqp.Language_EN, Label: "en-label"}}},
		"order": Order{ID: "bonta", Labels: []OrderLabel{
			{Locale: amqp.Language_FR, Label: "fr-label"}, {Locale: amqp.Language_EN, Label: "en-label"}}},
		"server": Server{ID: "bonta", Labels: []ServerLabel{
			{Locale: amqp.Language_FR, Label: "fr-label"}, {Locale: amqp.Language_EN, Label: "en-label"}}},
		"weapon area effect": WeaponAreaEffect{ID: "bonta", Labels: []WeaponAreaEffectLabel{
			{Locale: amqp.Language_FR, Label: "fr-label"}, {Locale: amqp.Language_EN, Label: "en-label"}}},
	}
	for name, entity := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, "bonta", entity.GetID())
			assert.Equal(t, expected, entity.GetLabels())
		})
	}
}

func TestLabelledEntities_WithoutLabelsReturnEmptyNonNilMap(t *testing.T) {
	for name, entity := range map[string]LabelledEntity{
		"city": City{}, "feed type": FeedType{}, "job": Job{},
		"order": Order{}, "server": Server{}, "weapon area effect": WeaponAreaEffect{},
	} {
		t.Run(name, func(t *testing.T) {
			labels := entity.GetLabels()
			assert.NotNil(t, labels)
			assert.Empty(t, labels)
		})
	}
}

func TestLabelledEntities_DuplicateLocaleKeepsLastLabel(t *testing.T) {
	server := Server{Labels: []ServerLabel{
		{Locale: amqp.Language_FR, Label: "first"},
		{Locale: amqp.Language_FR, Label: "second"},
	}}

	assert.Equal(t, labels(amqp.Language_FR, "second"), server.GetLabels())
}

func TestTwitterAccount_SameNameForEverySupportedLanguage(t *testing.T) {
	account := TwitterAccount{ID: "tw1", Name: "DOFUSfr"}

	assert.Equal(t, "tw1", account.GetID())
	assert.Equal(t, labels(
		amqp.Language_FR, "DOFUSfr",
		amqp.Language_EN, "DOFUSfr",
		amqp.Language_ES, "DOFUSfr",
		amqp.Language_DE, "DOFUSfr",
		amqp.Language_PT, "DOFUSfr",
	), account.GetLabels(), "every supported language, never ANY")
}

func TestGetNeutralCity_HasNeutralColorAndIconButNoIdentity(t *testing.T) {
	city := GetNeutralCity()

	assert.Equal(t, constants.NeutralCityColor, city.Color)
	assert.Contains(t, city.Icon, "neutral")
	assert.Empty(t, city.ID)
	assert.Empty(t, city.GetLabels())
}
