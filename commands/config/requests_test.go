package config

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	almanaxNewsChannel = "news-almanax-fr"
	webhookFollowPath  = "/api/v9/channels/" + almanaxNewsChannel + "/followers"
)

// webhookReadySession returns a session where the current channel is a text
// channel the bot may manage webhooks in.
func webhookReadySession(t *testing.T) (*discordgo.Session, *testutils.FakeDiscord) {
	t.Helper()
	s, fake := testutils.NewDiscordSession(t)
	addGuild(t, s, discordgo.PermissionManageWebhooks,
		textChannel(currentChannel), textChannel(otherChannel))
	return s, fake
}

func withAlmanaxNews(f *fixture) {
	f.almanax.newsLocale = discordgo.French
	f.almanax.news = &entities.AlmanaxNews{Locale: amqp.Language_FR, NewsChannelID: almanaxNewsChannel}
}

func singleRequest(t *testing.T, f *fixture) requestCall {
	t.Helper()
	require.Len(t, f.manager.calls, 1)
	call := f.manager.calls[0]
	assert.Equal(t, constants.ConfigurationRequestRoutingKey, call.routingKey)
	assert.Equal(t, authorID, call.message.UserID)
	assert.Equal(t, amqp.Language_FR, call.message.Language)
	assert.Equal(t, constants.GetGame().AMQPGame, call.message.Game)
	return call
}

// ---- get ----

func TestHandle_GetRequestsGuildConfiguration(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, configCommand(contract.ConfigGetSubCommandName))

	call := singleRequest(t, f)
	assert.Equal(t, amqp.RabbitMQMessage_CONFIGURATION_GET_REQUEST, call.message.Type)
	assert.Equal(t, guildID, call.message.GetConfigurationGetRequest().GetGuildId())
	assert.Empty(t, call.properties)
	assert.Empty(t, fake.Requests(), "the answer is sent by the callback only")
}

func TestHandle_RequestFailurePanics(t *testing.T) {
	f := newFixture(t)
	f.manager.err = errors.New("broker down")
	s, _ := testutils.NewDiscordSession(t)

	assert.PanicsWithError(t, "broker down", func() {
		f.cmd.Handle(s, configCommand(contract.ConfigGetSubCommandName))
	})
}

func TestHandle_MessageComponentsAreIgnored(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	assert.NotPanics(t, func() {
		f.cmd.Handle(s, interaction(discordgo.InteractionMessageComponent,
			discordgo.MessageComponentInteractionData{CustomID: "/config"}))
	})
	assert.Empty(t, f.manager.calls)
	assert.Empty(t, fake.Requests())
}

// ---- server ----

func TestHandle_ServerBindsResolvedServerToChannel(t *testing.T) {
	f := newFixture(t)
	f.servers.found = []entities.Server{server("djaul", "Djaul", "")}
	s, fake := webhookReadySession(t)

	f.cmd.Handle(s, configCommand(contract.ConfigServerSubCommandName,
		stringOption(contract.ConfigServerOptionName, "dja"), otherChannelOption()))

	call := singleRequest(t, f)
	assert.Equal(t, amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_REQUEST, call.message.Type)
	req := call.message.GetConfigurationSetServerRequest()
	assert.Equal(t, guildID, req.GetGuildId())
	assert.Equal(t, otherChannel, req.GetChannelId())
	assert.Equal(t, "djaul", req.GetServerId())
	assert.Empty(t, fake.Requests())
}

func TestHandle_ServerWithoutChannelBindsWholeGuild(t *testing.T) {
	f := newFixture(t)
	f.servers.found = []entities.Server{server("djaul", "Djaul", "")}
	s, _ := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, configCommand(contract.ConfigServerSubCommandName,
		stringOption(contract.ConfigServerOptionName, "dja")))

	req := singleRequest(t, f).message.GetConfigurationSetServerRequest()
	assert.Empty(t, req.GetChannelId())
	assert.Equal(t, "djaul", req.GetServerId())
}

func TestHandle_ServerUnknownAnswersErrorWithoutRequest(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, configCommand(contract.ConfigServerSubCommandName,
		stringOption(contract.ConfigServerOptionName, "nope")))

	assert.Empty(t, f.manager.calls)
	assert.Contains(t, editContent(t, fake), "'*nope*'")
}

func TestHandle_ServerWithoutServerOptionPanics(t *testing.T) {
	f := newFixture(t)
	s, _ := testutils.NewDiscordSession(t)

	assert.Panics(t, func() {
		f.cmd.Handle(s, configCommand(contract.ConfigServerSubCommandName, otherChannelOption()))
	})
	assert.Empty(t, f.manager.calls)
}

// ---- almanax ----

func TestHandle_AlmanaxDisableSkipsFollowAndSendsEmptyWebhook(t *testing.T) {
	f := newFixture(t)
	withAlmanaxNews(f)
	s, fake := webhookReadySession(t)

	f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName,
		enabledOption(false)))

	call := singleRequest(t, f)
	assert.Equal(t, amqp.RabbitMQMessage_CONFIGURATION_SET_NOTIFICATION_REQUEST, call.message.Type)
	req := call.message.GetConfigurationSetNotificationRequest()
	assert.Equal(t, guildID, req.GetGuildId())
	assert.Equal(t, currentChannel, req.GetChannelId())
	assert.Empty(t, req.GetWebhookId())
	assert.Empty(t, req.GetLabel())
	assert.Equal(t, amqp.NotificationType_ALMANAX, req.GetNotificationType())
	assert.False(t, req.GetEnabled())
	assert.Empty(t, fake.Requests(), "disabling never follows the announcement channel")
	assert.Equal(t, []discordgo.Locale{discordgo.French}, f.almanax.locales)
}

func TestHandle_AlmanaxEnableFollowsNewsChannelIntoTarget(t *testing.T) {
	f := newFixture(t)
	withAlmanaxNews(f)
	s, fake := webhookReadySession(t)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/followers", Status: 200,
		Body: `{"channel_id":"` + almanaxNewsChannel + `","webhook_id":"wh-42"}`})

	f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName,
		enabledOption(true), otherChannelOption()))

	follow := fake.Last(t)
	assert.Equal(t, "POST", follow.Method)
	assert.Equal(t, webhookFollowPath, follow.Path)
	var body struct {
		WebhookChannelID string `json:"webhook_channel_id"`
	}
	follow.Decode(t, &body)
	assert.Equal(t, otherChannel, body.WebhookChannelID)

	req := singleRequest(t, f).message.GetConfigurationSetNotificationRequest()
	assert.Equal(t, "wh-42", req.GetWebhookId())
	assert.Equal(t, otherChannel, req.GetChannelId())
	assert.True(t, req.GetEnabled())
}

func TestHandle_AlmanaxTooManyWebhooksAnswersErrorWithoutRequest(t *testing.T) {
	f := newFixture(t)
	withAlmanaxNews(f)
	s, fake := webhookReadySession(t)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/followers", Status: 400,
		Body: `{"code":30007,"message":"Maximum number of webhooks reached (15)"}`})

	f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName,
		enabledOption(true)))

	assert.Empty(t, f.manager.calls)
	assertEditedWebhook(t, fake)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "errors.too_many_webhooks")[0],
		editContent(t, fake))
}

func TestHandle_AlmanaxOtherFollowFailurePanics(t *testing.T) {
	f := newFixture(t)
	withAlmanaxNews(f)
	s, fake := webhookReadySession(t)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/followers", Status: 403,
		Body: `{"code":50013,"message":"Missing Permissions"}`})

	assert.Panics(t, func() {
		f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName,
			enabledOption(true)))
	})
	assert.Empty(t, f.manager.calls)
}

func TestHandle_AlmanaxWithoutNewsForLocalePanics(t *testing.T) {
	f := newFixture(t)
	s, _ := webhookReadySession(t)

	assert.PanicsWithError(t, "cannot find almanax news for locale French", func() {
		f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName,
			enabledOption(false)))
	})
	assert.Empty(t, f.manager.calls)
}

func TestHandle_AlmanaxStopsWhenTargetChannelIsNotAllowed(t *testing.T) {
	f := newFixture(t)
	withAlmanaxNews(f)
	s, fake := testutils.NewDiscordSession(t)
	addGuild(t, s, 0, textChannel(currentChannel))

	f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName,
		enabledOption(true)))

	assert.Empty(t, f.manager.calls)
	assert.Empty(t, f.almanax.locales)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "checks.webhook.permission")[0],
		editContent(t, fake))
}

// ---- rss ----

func frenchNewsFeed() entities.FeedType {
	return feedType("news", "Actualités",
		entities.FeedSource{Locale: amqp.Language_EN, FeedTypeID: "news", NewsChannelID: "news-rss-en"},
		entities.FeedSource{Locale: amqp.Language_FR, FeedTypeID: "news", NewsChannelID: "news-rss-fr"},
	)
}

func TestHandle_RSSFollowsSourceMatchingCallerLocale(t *testing.T) {
	f := newFixture(t)
	f.feeds.found = []entities.FeedType{frenchNewsFeed()}
	s, fake := webhookReadySession(t)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/followers", Status: 200,
		Body: `{"webhook_id":"wh-rss"}`})

	f.cmd.Handle(s, configCommand(contract.ConfigRSSSubCommandName,
		stringOption(contract.ConfigFeedTypeOptionName, "actu"),
		enabledOption(true)))

	assert.Equal(t, "/api/v9/channels/news-rss-fr/followers", fake.Last(t).Path)
	req := singleRequest(t, f).message.GetConfigurationSetNotificationRequest()
	assert.Equal(t, amqp.NotificationType_RSS, req.GetNotificationType())
	assert.Equal(t, "news", req.GetLabel(), "the feed type ID is the label")
	assert.Equal(t, "wh-rss", req.GetWebhookId())
	assert.Equal(t, currentChannel, req.GetChannelId())
	assert.True(t, req.GetEnabled())
}

func TestHandle_RSSDisableDoesNotFollow(t *testing.T) {
	f := newFixture(t)
	f.feeds.found = []entities.FeedType{frenchNewsFeed()}
	s, fake := webhookReadySession(t)

	f.cmd.Handle(s, configCommand(contract.ConfigRSSSubCommandName,
		stringOption(contract.ConfigFeedTypeOptionName, "actu"),
		enabledOption(false)))

	assert.Empty(t, fake.Requests())
	req := singleRequest(t, f).message.GetConfigurationSetNotificationRequest()
	assert.Empty(t, req.GetWebhookId())
	assert.False(t, req.GetEnabled())
}

func TestHandle_RSSWithoutSourceForLocaleFollowsDefaultLocaleSource(t *testing.T) {
	// Italian and the other locales the bot does not translate have no source
	// of their own; they get the default language instead of an error.
	f := newFixture(t)
	f.feeds.found = []entities.FeedType{feedType("news", "Actualités",
		entities.FeedSource{Locale: amqp.Language_EN, FeedTypeID: "news", NewsChannelID: "news-rss-en"})}
	s, fake := webhookReadySession(t)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/followers", Status: 200,
		Body: `{"webhook_id":"wh-rss-en"}`})

	f.cmd.Handle(s, configCommand(contract.ConfigRSSSubCommandName,
		stringOption(contract.ConfigFeedTypeOptionName, "actu"),
		enabledOption(true)))

	assert.Equal(t, "/api/v9/channels/news-rss-en/followers", fake.Last(t).Path)
	req := singleRequest(t, f).message.GetConfigurationSetNotificationRequest()
	assert.Equal(t, "wh-rss-en", req.GetWebhookId())
	assert.True(t, req.GetEnabled())
}

func TestHandle_RSSWithoutAnySourcePanics(t *testing.T) {
	f := newFixture(t)
	f.feeds.found = []entities.FeedType{feedType("news", "Actualités")}
	s, _ := webhookReadySession(t)

	assert.PanicsWithError(t, "cannot find feed source for 'news' in French", func() {
		f.cmd.Handle(s, configCommand(contract.ConfigRSSSubCommandName,
			stringOption(contract.ConfigFeedTypeOptionName, "actu"),
			enabledOption(false)))
	})
	assert.Empty(t, f.manager.calls)
}

func TestHandle_AlmanaxWithoutNewsForLocaleFollowsDefaultLocaleNews(t *testing.T) {
	f := newFixture(t)
	f.almanax.newsLocale = i18n.DefaultLocale
	f.almanax.news = &entities.AlmanaxNews{Locale: amqp.Language_EN, NewsChannelID: "news-almanax-en"}
	s, fake := webhookReadySession(t)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/followers", Status: 200,
		Body: `{"webhook_id":"wh-almanax-en"}`})

	f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName, enabledOption(true)))

	assert.Equal(t, "/api/v9/channels/news-almanax-en/followers", fake.Last(t).Path)
	assert.Equal(t, []discordgo.Locale{discordgo.French, i18n.DefaultLocale}, f.almanax.locales,
		"the caller's locale is tried first, then the default one")
	req := singleRequest(t, f).message.GetConfigurationSetNotificationRequest()
	assert.Equal(t, "wh-almanax-en", req.GetWebhookId())
	assert.True(t, req.GetEnabled())
}

func TestHandle_AlmanaxWithoutNewsAtAllPanics(t *testing.T) {
	f := newFixture(t)
	s, _ := webhookReadySession(t)

	assert.PanicsWithError(t, "cannot find almanax news for locale French", func() {
		f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName, enabledOption(true)))
	})
	assert.Empty(t, f.manager.calls)
}

func TestHandle_RSSUnknownFeedStopsBeforeChannelChecks(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t) // empty state: channel checks would panic

	f.cmd.Handle(s, configCommand(contract.ConfigRSSSubCommandName,
		stringOption(contract.ConfigFeedTypeOptionName, "nope"),
		enabledOption(true)))

	assert.Empty(t, f.manager.calls)
	assert.Contains(t, editContent(t, fake), "'*nope*'")
}

// ---- twitter ----

func TestHandle_TwitterFollowsAccountNewsChannel(t *testing.T) {
	f := newFixture(t)
	f.twitters.found = []entities.TwitterAccount{{ID: "dofusfr", Name: "DOFUSfr", NewsChannelID: "news-tw"}}
	s, fake := webhookReadySession(t)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/followers", Status: 200,
		Body: `{"webhook_id":"wh-tw"}`})

	f.cmd.Handle(s, configCommand(contract.ConfigTwitterSubCommandName,
		stringOption(contract.ConfigTwitterAccountOptionName, "dofus"),
		enabledOption(true)))

	assert.Equal(t, "/api/v9/channels/news-tw/followers", fake.Last(t).Path)
	req := singleRequest(t, f).message.GetConfigurationSetNotificationRequest()
	assert.Equal(t, amqp.NotificationType_TWITTER, req.GetNotificationType())
	assert.Equal(t, "dofusfr", req.GetLabel())
	assert.Equal(t, "wh-tw", req.GetWebhookId())
	assert.True(t, req.GetEnabled())
}

func TestHandle_TwitterTooManyWebhooksAnswersError(t *testing.T) {
	f := newFixture(t)
	f.twitters.found = []entities.TwitterAccount{{ID: "dofusfr", Name: "DOFUSfr", NewsChannelID: "news-tw"}}
	s, fake := webhookReadySession(t)
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/followers", Status: 400,
		Body: `{"code":30007}`})

	f.cmd.Handle(s, configCommand(contract.ConfigTwitterSubCommandName,
		stringOption(contract.ConfigTwitterAccountOptionName, "dofus"),
		enabledOption(true)))

	assert.Empty(t, f.manager.calls)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "errors.too_many_webhooks")[0],
		editContent(t, fake))
}

func TestHandle_TwitterDisableDoesNotFollow(t *testing.T) {
	f := newFixture(t)
	f.twitters.found = []entities.TwitterAccount{{ID: "dofusfr", Name: "DOFUSfr", NewsChannelID: "news-tw"}}
	s, fake := webhookReadySession(t)

	f.cmd.Handle(s, configCommand(contract.ConfigTwitterSubCommandName,
		stringOption(contract.ConfigTwitterAccountOptionName, "dofus"),
		enabledOption(false)))

	assert.Empty(t, fake.Requests())
	req := singleRequest(t, f).message.GetConfigurationSetNotificationRequest()
	assert.Empty(t, req.GetWebhookId())
	assert.False(t, req.GetEnabled())
}

// ---- option extraction ----

func TestGetOptions_WrongContextTypesAreErrors(t *testing.T) {
	ctx := context.Background()
	_, _, err := getServerOptions(ctx)
	require.Error(t, err)
	_, _, err = getServerOptions(context.WithValue(
		context.WithValue(ctx, constants.ContextKeyServer, entities.Server{}), constants.ContextKeyChannel, 42))
	require.Error(t, err, "a non-string channel is rejected")

	withChannel := channelCtx("c")
	_, _, err = getWebhookAlmanaxOptions(ctx)
	require.Error(t, err)
	_, _, err = getWebhookAlmanaxOptions(withChannel)
	require.Error(t, err, "enabled is mandatory")

	_, _, _, err = getWebhookTwitterOptions(ctx)
	require.Error(t, err)
	_, _, _, err = getWebhookTwitterOptions(withChannel)
	require.Error(t, err)
	_, _, _, err = getWebhookTwitterOptions(
		context.WithValue(withChannel, constants.ContextKeyTwitter, entities.TwitterAccount{}))
	require.Error(t, err)

	_, _, _, err = getWebhookRssOptions(ctx)
	require.Error(t, err)
	_, _, _, err = getWebhookRssOptions(withChannel)
	require.Error(t, err)
	_, _, _, err = getWebhookRssOptions(context.WithValue(withChannel, constants.ContextKeyFeed, entities.FeedType{}))
	require.Error(t, err)
}

// ---- answers ----

func setServerAnswer(status amqp.RabbitMQMessage_Status, messageType amqp.RabbitMQMessage_Type,
) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{Type: messageType, Status: status, Language: amqp.Language_EN}
}

func TestServerRespond_SuccessEditsWithAnswerLanguage(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.serverRespond(context.Background(), s, configCommand("server"),
		setServerAnswer(amqp.RabbitMQMessage_SUCCESS, amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_ANSWER), nil)

	assertEditedWebhook(t, fake)
	assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "config.success")[0], editContent(t, fake),
		"the answer language wins over the interaction locale")
}

func TestServerRespond_InvalidAnswerPanics(t *testing.T) {
	cases := map[string]*amqp.RabbitMQMessage{
		"failed status": setServerAnswer(amqp.RabbitMQMessage_FAILED,
			amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_ANSWER),
		"wrong type": setServerAnswer(amqp.RabbitMQMessage_SUCCESS,
			amqp.RabbitMQMessage_CONFIGURATION_SET_NOTIFICATION_ANSWER),
	}
	for name, message := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)

			assert.PanicsWithError(t, commands.ErrInvalidAnswerMessage.Error(), func() {
				f.cmd.serverRespond(context.Background(), s, configCommand("server"), message, nil)
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func notificationAnswer(removeWebhook bool) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_CONFIGURATION_SET_NOTIFICATION_ANSWER,
		Status:   amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_FR,
		ConfigurationSetNotificationAnswer: &amqp.ConfigurationSetNotificationAnswer{
			WebhookId: "wh-old", RemoveWebhook: removeWebhook,
		},
	}
}

func TestSetNotificationRespond_KeepWebhookOnlyEdits(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.setNotificationRespond(context.Background(), s, configCommand("almanax"), notificationAnswer(false), nil)

	require.Len(t, fake.Requests(), 1)
	assertEditedWebhook(t, fake)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "config.success")[0], editContent(t, fake))
}

func TestSetNotificationRespond_RemoveWebhookDeletesItBeforeEditing(t *testing.T) {
	for name, deleteStatus := range map[string]int{
		"deleted":         204,
		"already deleted": 404,
		"delete failure":  403,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)
			fake.Respond(testutils.DiscordResponse{Method: "DELETE", PathContains: "/webhooks/wh-old",
				Status: deleteStatus, Body: `{"code":0}`})

			f.cmd.setNotificationRespond(context.Background(), s, configCommand("almanax"),
				notificationAnswer(true), nil)

			requests := fake.Requests()
			require.Len(t, requests, 2)
			assert.Equal(t, "DELETE", requests[0].Method)
			assert.Equal(t, "/api/v9/webhooks/wh-old", requests[0].Path)
			assertEditedWebhook(t, fake)
			assert.Equal(t, testutils.Translations(t, discordgo.French, "config.success")[0],
				editContent(t, fake), "deletion errors never block the confirmation")
		})
	}
}

func TestSetNotificationRespond_InvalidAnswerPanics(t *testing.T) {
	failed := notificationAnswer(false)
	failed.Status = amqp.RabbitMQMessage_FAILED
	wrongType := notificationAnswer(false)
	wrongType.Type = amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_ANSWER
	missingAnswer := notificationAnswer(false)
	missingAnswer.ConfigurationSetNotificationAnswer = nil

	for name, message := range map[string]*amqp.RabbitMQMessage{
		"failed status": failed, "wrong type": wrongType, "missing answer": missingAnswer,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)

			assert.PanicsWithError(t, commands.ErrInvalidAnswerMessage.Error(), func() {
				f.cmd.setNotificationRespond(context.Background(), s, configCommand("almanax"), message, nil)
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func configGetAnswer(answer *amqp.ConfigurationGetAnswer) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:                   amqp.RabbitMQMessage_CONFIGURATION_GET_ANSWER,
		Status:                 amqp.RabbitMQMessage_SUCCESS,
		Language:               amqp.Language_FR,
		ConfigurationGetAnswer: answer,
	}
}

func TestGetRespond_BuildsEmbedFromGuildAndValidChannels(t *testing.T) {
	f := newFixture(t)
	f.servers.byID = map[string]entities.Server{"djaul": server("djaul", "Djaul", "")}
	f.feeds.byID = map[string]*entities.FeedType{"news": {ID: "news", Labels: []entities.FeedTypeLabel{
		{Locale: amqp.Language_FR, FeedTypeID: "news", Label: "Actualités"}}}}
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/guilds/" + guildID, Status: 200,
		Body: `{"id":"` + guildID + `","name":"Les Kaellistes","icon":"abc"}`})
	for _, id := range []string{"ch-a", "ch-b"} {
		fake.Respond(testutils.DiscordResponse{PathContains: "/channels/" + id, Status: 200,
			Body: `{"id":"` + id + `","guild_id":"` + guildID + `"}`})
	}
	fake.Respond(testutils.DiscordResponse{PathContains: "/channels/ch-gone", Status: 404, Body: `{"code":10003}`})
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/wh-ok", Status: 200,
		Body: `{"id":"wh-ok","channel_id":"ch-a","guild_id":"` + guildID + `"}`})
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/wh-moved", Status: 200,
		Body: `{"id":"wh-moved","channel_id":"ch-elsewhere","guild_id":"` + guildID + `"}`})
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/wh-deleted", Status: 404, Body: `{"code":10015}`})

	f.cmd.getRespond(context.Background(), s, configCommand("get"), configGetAnswer(&amqp.ConfigurationGetAnswer{
		GuildId:  guildID,
		ServerId: "djaul",
		ServerChannels: []*amqp.ConfigurationGetAnswer_ServerChannel{
			{ChannelId: "ch-a", ServerId: "djaul"},
			{ChannelId: "ch-gone", ServerId: "djaul"},
		},
		NotifiedChannels: []*amqp.ConfigurationGetAnswer_NotifiedChannel{
			{ChannelId: "ch-a", WebhookId: "wh-ok", NotificationType: amqp.NotificationType_ALMANAX},
			{ChannelId: "ch-b", WebhookId: "wh-moved", NotificationType: amqp.NotificationType_ALMANAX},
			{ChannelId: "ch-b", WebhookId: "wh-deleted", Label: "news", NotificationType: amqp.NotificationType_RSS},
		},
	}), nil)

	assertEditedWebhook(t, fake)
	payload := fake.LastEdit(t)
	assert.Nil(t, payload.Content)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "Les Kaellistes", embed.Title)
	assert.Equal(t, discordgo.EndpointGuildIcon(guildID, "abc")+"?size=128", embed.Thumbnail.URL)
	assert.Contains(t, embed.Description, "**Djaul**")
	require.Len(t, embed.Fields, 2)
	assert.Equal(t, "<#ch-a> :djaul: **Djaul**\n", embed.Fields[0].Value,
		"channels Discord cannot resolve are dropped")
	assert.Equal(t, "<#ch-a> :almanax: **Almanax**\n", embed.Fields[1].Value,
		"webhooks that moved or vanished are dropped")

	channelFetches := 0
	for _, req := range fake.Requests() {
		if req.Path == "/api/v9/channels/ch-a" {
			channelFetches++
		}
	}
	assert.Equal(t, 1, channelFetches, "channels are fetched once and cached")
}

func TestGetRespond_EmptyConfiguration(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/guilds/", Status: 200,
		Body: `{"id":"` + guildID + `","name":"Nouvelle guilde"}`})

	f.cmd.getRespond(context.Background(), s, configCommand("get"),
		configGetAnswer(&amqp.ConfigurationGetAnswer{GuildId: guildID}), nil)

	embed := fake.LastEdit(t).Embeds[0]
	assert.Equal(t, "Nouvelle guilde", embed.Title)
	assert.Empty(t, embed.Thumbnail.URL, "no icon means no thumbnail")
	assert.True(t, strings.HasPrefix(embed.Description, "Aucun serveur"), embed.Description)
	assert.Len(t, fake.Requests(), 2, "guild fetch then edit")
}

func TestGetRespond_GuildFetchFailurePanics(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/guilds/", Status: 403, Body: `{"code":50001}`})

	assert.Panics(t, func() {
		f.cmd.getRespond(context.Background(), s, configCommand("get"),
			configGetAnswer(&amqp.ConfigurationGetAnswer{GuildId: guildID}), nil)
	})
	assert.Len(t, fake.Requests(), 1, "no edit is attempted")
}

func TestGetRespond_FailedStatusPanics(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	message := configGetAnswer(nil)
	message.Status = amqp.RabbitMQMessage_FAILED

	assert.PanicsWithError(t, commands.ErrInvalidAnswerMessage.Error(), func() {
		f.cmd.getRespond(context.Background(), s, configCommand("get"), message, nil)
	})
	assert.Empty(t, fake.Requests())
}

func TestHandle_CallbacksWiredToMatchingResponders(t *testing.T) {
	f := newFixture(t)
	withAlmanaxNews(f)
	f.servers.found = []entities.Server{server("djaul", "Djaul", "")}
	s, fake := webhookReadySession(t)

	f.cmd.Handle(s, configCommand(contract.ConfigServerSubCommandName,
		stringOption(contract.ConfigServerOptionName, "dja")))
	f.cmd.Handle(s, configCommand(contract.ConfigAlmanaxSubCommandName,
		enabledOption(false)))
	require.Len(t, f.manager.calls, 2)

	// server callback only accepts server answers
	assert.PanicsWithError(t, commands.ErrInvalidAnswerMessage.Error(), func() {
		f.manager.calls[0].callback(context.Background(), s, configCommand("server"), notificationAnswer(false), nil)
	})
	f.manager.calls[0].callback(context.Background(), s, configCommand("server"),
		setServerAnswer(amqp.RabbitMQMessage_SUCCESS, amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_ANSWER), nil)
	assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "config.success")[0], editContent(t, fake))

	// notification callback only accepts notification answers
	f.manager.calls[1].callback(context.Background(), s, configCommand("almanax"), notificationAnswer(false), nil)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "config.success")[0], editContent(t, fake))
}
