package mappers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func channel(id string) *discordgo.Channel { return &discordgo.Channel{ID: id} }

func configServices() (fakeServers, fakeFeeds, fakeTwitters) {
	return fakeServers{servers: []entities.Server{testServer()}},
		fakeFeeds{feedTypes: []entities.FeedType{{ID: "devblog", Labels: []entities.FeedTypeLabel{
			{FeedTypeID: "devblog", Locale: amqp.Language_EN, Label: "Devblog"},
			{FeedTypeID: "devblog", Locale: amqp.Language_FR, Label: "Devblog-FR"},
		}}}},
		fakeTwitters{accounts: []entities.TwitterAccount{{ID: "tw", Name: "DOFUS_EN"}}}
}

func TestMapConfigurationRequests_CopyFields(t *testing.T) {
	get := MapConfigurationGetRequest("guild", "author", discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_CONFIGURATION_GET_REQUEST, get.Type)
	assert.Equal(t, amqp.Language_FR, get.Language)
	assert.Equal(t, "author", get.UserID)
	require.NotNil(t, get.ConfigurationGetRequest)
	assert.Equal(t, "guild", get.ConfigurationGetRequest.GuildId)

	server := MapConfigurationServerRequest("guild", "chan", "srv", "author", discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_REQUEST, server.Type)
	assert.Equal(t, "author", server.UserID)
	require.NotNil(t, server.ConfigurationSetServerRequest)
	assert.Equal(t, "guild", server.ConfigurationSetServerRequest.GuildId)
	assert.Equal(t, "chan", server.ConfigurationSetServerRequest.ChannelId)
	assert.Equal(t, "srv", server.ConfigurationSetServerRequest.ServerId)

	notif := MapConfigurationNotificationRequest("guild", "chan", "hook", "author", "devblog",
		amqp.NotificationType_RSS, true, discordgo.German)
	assert.Equal(t, amqp.RabbitMQMessage_CONFIGURATION_SET_NOTIFICATION_REQUEST, notif.Type)
	assert.Equal(t, amqp.Language_DE, notif.Language)
	assert.Equal(t, "author", notif.UserID)
	require.NotNil(t, notif.ConfigurationSetNotificationRequest)
	payload := notif.ConfigurationSetNotificationRequest
	assert.Equal(t, "guild", payload.GuildId)
	assert.Equal(t, "chan", payload.ChannelId)
	assert.Equal(t, "hook", payload.WebhookId)
	assert.Equal(t, "devblog", payload.Label)
	assert.Equal(t, amqp.NotificationType_RSS, payload.NotificationType)
	assert.True(t, payload.Enabled)
}

func TestMapNotifiedChannelsToI18n_ResolvesProvidersAndSkipsUnknown(t *testing.T) {
	_, feedService, twitterService := configServices()
	webhooks := []constants.NotifiedChannel{
		{Channel: channel("c1"), NotificationType: amqp.NotificationType_ALMANAX},
		{Channel: channel("c2"), NotificationType: amqp.NotificationType_RSS, Label: "devblog"},
		{Channel: channel("c3"), NotificationType: amqp.NotificationType_RSS, Label: "missing-feed"},
		{Channel: channel("c4"), NotificationType: amqp.NotificationType_TWITTER, Label: "tw"},
		{Channel: channel("c5"), NotificationType: amqp.NotificationType_TWITTER, Label: "missing-tw"},
		{Channel: channel("c6"), NotificationType: amqp.NotificationType_UNKNOWN},
	}

	result := mapNotifiedChannelsToI18n(webhooks, fakeEmojis{}, feedService, twitterService, discordgo.French)

	require.Len(t, result, 3)
	assert.Equal(t, "<#c1>", result[0].Channel)
	assert.Equal(t, "<misc:almanax>", result[0].Provider.Emoji)
	assertTranslated(t, discordgo.French, "config.embed.webhook.almanax", result[0].Provider.Name)
	assert.Equal(t, i18nChannelWebhook{Channel: "<#c2>",
		Provider: i18nProvider{Name: "Devblog-FR", Emoji: "<misc:rss>"}}, result[1])
	assert.Equal(t, i18nChannelWebhook{Channel: "<#c4>",
		Provider: i18nProvider{Name: "DOFUS_EN", Emoji: "<misc:twitter>"}}, result[2])
}

func TestMapNotifiedChannelsToI18n_NoWebhooksGivesEmptySlice(t *testing.T) {
	_, feedService, twitterService := configServices()
	result := mapNotifiedChannelsToI18n(nil, fakeEmojis{}, feedService, twitterService, discordgo.French)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestMapConfigToEmbed_ListsGuildServerChannelsAndWebhooks(t *testing.T) {
	serverService, feedService, twitterService := configServices()
	guild := constants.GuildConfig{
		Name:     "Les Fous",
		Icon:     "guild.png",
		ServerID: "srv",
		ServerChannels: []constants.ServerChannel{
			{Channel: channel("c1"), ServerID: "srv"},
			{Channel: channel("c2"), ServerID: "ghost-srv"},
		},
		NotifiedChannels: []constants.NotifiedChannel{
			{Channel: channel("c3"), NotificationType: amqp.NotificationType_RSS, Label: "devblog"},
		},
	}

	embed := MapConfigToEmbed(guild, fakeEmojis{}, serverService, feedService, twitterService, amqp.Language_EN)

	assert.Equal(t, "Les Fous", embed.Title)
	assert.Equal(t, "guild.png", embed.Thumbnail.URL)
	assert.Equal(t, constants.Color, embed.Color)
	assertDefaultFooter(t, embed.Footer)
	assertRendered(t, embed.Description)
	assert.Contains(t, embed.Description, "<Server:srv>")
	assert.Contains(t, embed.Description, "**Imagiro**")

	require.Len(t, embed.Fields, 2)
	serverField, webhookField := embed.Fields[0], embed.Fields[1]
	assert.Contains(t, serverField.Name, "<misc:dofusGame>")
	assertInOrder(t, serverField.Value,
		"<#c1> <Server:srv> **Imagiro**",
		"<#c2> <Server:ghost-srv> **ghost-srv**")
	assertTranslated(t, discordgo.EnglishGB, "config.embed.webhook.name", webhookField.Name)
	assert.Contains(t, webhookField.Value, "<#c3> <misc:rss> **Devblog**")
	assert.False(t, serverField.Inline)
	assert.False(t, webhookField.Inline)
}

func TestMapConfigToEmbed_NoServerNorChannels(t *testing.T) {
	serverService, feedService, twitterService := configServices()

	embed := MapConfigToEmbed(constants.GuildConfig{Name: "Empty"}, fakeEmojis{},
		serverService, feedService, twitterService, amqp.Language_FR)

	assertRendered(t, embed.Description)
	assert.Contains(t, embed.Description, constants.GetGame().Name)
	assert.NotContains(t, embed.Description, "<Server:")
	for _, field := range embed.Fields {
		assertRendered(t, field.Value)
		assert.NotContains(t, field.Value, "<#")
	}
}

func TestMapConfigToEmbed_UnknownGuildServerFallsBackToItsID(t *testing.T) {
	serverService, feedService, twitterService := configServices()

	embed := MapConfigToEmbed(constants.GuildConfig{ServerID: "ghost"}, fakeEmojis{},
		serverService, feedService, twitterService, amqp.Language_EN)

	assert.Contains(t, embed.Description, "<Server:ghost> **ghost**")
}
