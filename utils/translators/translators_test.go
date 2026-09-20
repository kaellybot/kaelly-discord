package translators

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/stretchr/testify/assert"
)

type label struct {
	locale amqp.Language
	value  string
}

func server(id string, labels ...label) entities.Server {
	server := entities.Server{ID: id}
	for _, l := range labels {
		server.Labels = append(server.Labels, entities.ServerLabel{ServerID: id, Locale: l.locale, Label: l.value})
	}
	return server
}

func TestGetEntityLabel_FallbackChain(t *testing.T) {
	frDjaul := label{amqp.Language_FR, "Djaul FR"}
	enDjaul := label{amqp.Language_EN, "Djaul EN"}
	cases := []struct {
		name     string
		entity   entities.LabelledEntity
		locale   discordgo.Locale
		expected string
	}{
		{"locale label wins", server("djaul", frDjaul, enDjaul), discordgo.French, "Djaul FR"},
		{"US English maps to EN labels", server("djaul", frDjaul, enDjaul), discordgo.EnglishUS, "Djaul EN"},
		{"missing locale falls back to default locale", server("djaul", enDjaul), discordgo.German, "Djaul EN"},
		{"unsupported locale falls back to default locale", server("djaul", frDjaul, enDjaul),
			discordgo.Japanese, "Djaul EN"},
		{"no locale nor default label falls back to ID", server("djaul", frDjaul), discordgo.German, "djaul"},
		{"no labels at all falls back to ID", server("djaul"), discordgo.French, "djaul"},
		{"an empty label is still a label", server("djaul", label{amqp.Language_FR, ""}, enDjaul),
			discordgo.French, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, GetEntityLabel(tc.entity, tc.locale))
		})
	}
}

func TestGetLabels_PreserveOrderPerEntityKind(t *testing.T) {
	fr := amqp.Language_FR
	assert.Equal(t, []string{"Bonta", "brakmar"}, GetCitiesLabels([]entities.City{
		{ID: "bonta", Labels: []entities.CityLabel{{Locale: fr, Label: "Bonta"}}},
		{ID: "brakmar"},
	}, discordgo.French))
	assert.Equal(t, []string{"Esprit"}, GetOrdersLabels([]entities.Order{
		{ID: "spirit", Labels: []entities.OrderLabel{{Locale: fr, Label: "Esprit"}}},
	}, discordgo.French))
	assert.Equal(t, []string{"Actualités"}, GetFeedTypesLabels([]entities.FeedType{
		{ID: "news", Labels: []entities.FeedTypeLabel{{Locale: fr, Label: "Actualités"}}},
	}, discordgo.French))
	assert.Equal(t, []string{"Alchimiste", "Paysan"}, GetJobsLabels([]entities.Job{
		{ID: "alchemist", Labels: []entities.JobLabel{{Locale: fr, Label: "Alchimiste"}}},
		{ID: "farmer", Labels: []entities.JobLabel{{Locale: fr, Label: "Paysan"}}},
	}, discordgo.French))
	assert.Equal(t, []string{"Djaul", "Imagiro"}, GetServersLabels([]entities.Server{
		server("djaul", label{fr, "Djaul"}),
		server("imagiro", label{fr, "Imagiro"}),
	}, discordgo.French))
	assert.Equal(t, []string{"DOFUS_FR"}, GetTwittersLabels([]entities.TwitterAccount{
		{ID: "tw", Name: "DOFUS_FR"},
	}, discordgo.PortugueseBR))
}

func TestGetLabels_EmptyInputGivesEmptyNonNilSlice(t *testing.T) {
	for name, labels := range map[string][]string{
		"cities":     GetCitiesLabels(nil, discordgo.French),
		"orders":     GetOrdersLabels(nil, discordgo.French),
		"feed types": GetFeedTypesLabels(nil, discordgo.French),
		"jobs":       GetJobsLabels(nil, discordgo.French),
		"servers":    GetServersLabels(nil, discordgo.French),
		"twitters":   GetTwittersLabels(nil, discordgo.French),
	} {
		assert.NotNil(t, labels, name)
		assert.Empty(t, labels, name)
	}
}

func TestFormatNumber_UsesLocaleGrouping(t *testing.T) {
	cases := []struct {
		locale   discordgo.Locale
		value    int64
		expected string
	}{
		{discordgo.EnglishGB, 1234567, "1,234,567"},
		{discordgo.EnglishUS, 1234567, "1,234,567"},
		{discordgo.French, 1234567, "1 234 567"},
		{discordgo.German, 1234567, "1.234.567"},
		{discordgo.SpanishES, 1234567, "1.234.567"},
		{discordgo.PortugueseBR, 1234567, "1.234.567"},
		{discordgo.Japanese, 1234567, "1,234,567"},
		{discordgo.EnglishGB, -1000, "-1,000"},
		{discordgo.French, 999, "999"},
		{discordgo.EnglishGB, 0, "0"},
	}
	for _, tc := range cases {
		t.Run(string(tc.locale), func(t *testing.T) {
			assert.Equal(t, tc.expected, FormatNumber(tc.value, tc.locale))
		})
	}
}
