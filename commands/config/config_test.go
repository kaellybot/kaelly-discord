package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/services/almanaxes"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/twitters"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	guildID        = "guild-1"
	currentChannel = "chan-current"
	otherChannel   = "chan-other"
	authorID       = "author-1"
)

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

// ---- service doubles ----

type findCall struct {
	name   string
	locale discordgo.Locale
	limit  int
}

type fakeServers struct {
	servers.Service
	found []entities.Server
	byID  map[string]entities.Server
	calls []findCall
}

func (f *fakeServers) FindServers(name string, locale discordgo.Locale, limit int) []entities.Server {
	f.calls = append(f.calls, findCall{name, locale, limit})
	return f.found
}

func (f *fakeServers) GetServer(id string) (entities.Server, bool) {
	server, found := f.byID[id]
	return server, found
}

type fakeFeeds struct {
	feeds.Service
	found []entities.FeedType
	byID  map[string]*entities.FeedType
	calls []findCall
}

func (f *fakeFeeds) FindFeedTypes(name string, locale discordgo.Locale, limit int) []entities.FeedType {
	f.calls = append(f.calls, findCall{name, locale, limit})
	return f.found
}

func (f *fakeFeeds) GetFeedType(id string) *entities.FeedType {
	return f.byID[id]
}

type fakeTwitters struct {
	twitters.Service
	found []entities.TwitterAccount
	byID  map[string]*entities.TwitterAccount
	calls []findCall
}

func (f *fakeTwitters) FindTwitterAccounts(name string, locale discordgo.Locale,
	limit int) []entities.TwitterAccount {
	f.calls = append(f.calls, findCall{name, locale, limit})
	return f.found
}

func (f *fakeTwitters) GetTwitterAccount(id string) *entities.TwitterAccount {
	return f.byID[id]
}

type fakeAlmanaxes struct {
	almanaxes.Service
	newsLocale discordgo.Locale
	news       *entities.AlmanaxNews
	locales    []discordgo.Locale
}

func (f *fakeAlmanaxes) GetAlmanaxNews(locale discordgo.Locale) *entities.AlmanaxNews {
	f.locales = append(f.locales, locale)
	if locale != f.newsLocale {
		return nil
	}
	return f.news
}

type fakeGuilds struct {
	guilds.Service
}

type fakeEmojis struct {
	emojis.Service
}

func (fakeEmojis) GetMiscStringEmoji(emojiID constants.EmojiMiscID) string {
	return ":" + string(emojiID) + ":"
}

func (fakeEmojis) GetEntityStringEmoji(entityID string, _ constants.EmojiEntity) string {
	return ":" + entityID + ":"
}

type requestCall struct {
	routingKey string
	message    *amqp.RabbitMQMessage
	callback   requests.RequestCallback
	properties []map[string]any
}

type fakeRequestManager struct {
	requests.RequestManager
	err   error
	calls []requestCall
}

func (f *fakeRequestManager) Request(_ *discordgo.Session, _ *discordgo.InteractionCreate,
	routingKey string, message *amqp.RabbitMQMessage, callback requests.RequestCallback,
	properties ...map[string]any) error {
	f.calls = append(f.calls, requestCall{routingKey, message, callback, properties})
	return f.err
}

type fixture struct {
	cmd      *Command
	servers  *fakeServers
	feeds    *fakeFeeds
	twitters *fakeTwitters
	almanax  *fakeAlmanaxes
	manager  *fakeRequestManager
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	viper.Set(constants.ConfigID, "config-discord-id")
	t.Cleanup(func() { viper.Set(constants.ConfigID, nil) })
	f := &fixture{
		servers:  &fakeServers{},
		feeds:    &fakeFeeds{},
		twitters: &fakeTwitters{},
		almanax:  &fakeAlmanaxes{},
		manager:  &fakeRequestManager{},
	}
	f.cmd = New(f.almanax, fakeEmojis{}, f.feeds, fakeGuilds{}, f.servers, f.twitters, f.manager)
	return f
}

// ---- entities ----

func server(id, frLabel, enLabel string) entities.Server {
	var labels []entities.ServerLabel
	if frLabel != "" {
		labels = append(labels, entities.ServerLabel{Locale: amqp.Language_FR, ServerID: id, Label: frLabel})
	}
	if enLabel != "" {
		labels = append(labels, entities.ServerLabel{Locale: amqp.Language_EN, ServerID: id, Label: enLabel})
	}
	return entities.Server{ID: id, Labels: labels}
}

func feedType(id, frLabel string, sources ...entities.FeedSource) entities.FeedType {
	return entities.FeedType{
		ID:      id,
		Labels:  []entities.FeedTypeLabel{{Locale: amqp.Language_FR, FeedTypeID: id, Label: frLabel}},
		Sources: sources,
	}
}

// ---- interactions ----

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		GuildID: guildID, ChannelID: currentChannel,
		Locale: discordgo.French, Data: data,
		Member: &discordgo.Member{User: &discordgo.User{ID: authorID}},
	}}
}

func configCommand(subCommand string,
	options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand, discordgo.ApplicationCommandInteractionData{
		Name: contract.ConfigCommandName,
		Options: []*discordgo.ApplicationCommandInteractionDataOption{{
			Name: subCommand, Type: discordgo.ApplicationCommandOptionSubCommand, Options: options,
		}},
	})
}

func configAutocomplete(subCommand string,
	options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	event := configCommand(subCommand, options...)
	event.Type = discordgo.InteractionApplicationCommandAutocomplete
	return event
}

func stringOption(name, value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: name, Type: discordgo.ApplicationCommandOptionString, Value: value,
	}
}

func focused(option *discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.ApplicationCommandInteractionDataOption {
	option.Focused = true
	return option
}

func enabledOption(value bool) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: contract.ConfigEnabledOptionName, Type: discordgo.ApplicationCommandOptionBoolean, Value: value,
	}
}

func otherChannelOption() *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: contract.ConfigChannelOptionName, Type: discordgo.ApplicationCommandOptionChannel, Value: otherChannel,
	}
}

// addGuild registers the guild, the bot member and the given channels in the
// session state. The @everyone role carries botPermissions.
func addGuild(t *testing.T, s *discordgo.Session, botPermissions int64, channels ...*discordgo.Channel) {
	t.Helper()
	require.NoError(t, s.State.GuildAdd(&discordgo.Guild{
		ID: guildID, OwnerID: "someone-else",
		Roles: []*discordgo.Role{{ID: guildID, Permissions: botPermissions}},
	}))
	require.NoError(t, s.State.MemberAdd(&discordgo.Member{
		GuildID: guildID, User: &discordgo.User{ID: s.State.User.ID},
	}))
	for _, channel := range channels {
		channel.GuildID = guildID
		require.NoError(t, s.State.ChannelAdd(channel))
	}
}

func textChannel(id string) *discordgo.Channel {
	return &discordgo.Channel{ID: id, Type: discordgo.ChannelTypeGuildText}
}

// nextRecorder is a middleware "next" capturing the context it was called with.
type nextRecorder struct {
	called bool
	ctx    context.Context
}

func (n *nextRecorder) next(ctx context.Context) {
	n.called = true
	n.ctx = ctx
}

func editContent(t *testing.T, fake *testutils.FakeDiscord) string {
	t.Helper()
	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	return *payload.Content
}

func assertEditedWebhook(t *testing.T, fake *testutils.FakeDiscord) {
	t.Helper()
	req := fake.Last(t)
	assert.Equal(t, "PATCH", req.Method)
	assert.Equal(t, "/api/v9/webhooks/app/tok/messages/@original", req.Path)
}

// ---- Matches / descriptions ----

func TestMatches_OnlyConfigApplicationCommandsAndAutocompletes(t *testing.T) {
	cmd := newFixture(t).cmd
	other := interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: "job"})
	component := interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: "/config"})

	assert.True(t, cmd.Matches(configCommand(contract.ConfigGetSubCommandName)))
	assert.True(t, cmd.Matches(configAutocomplete(contract.ConfigServerSubCommandName)))
	assert.False(t, cmd.Matches(other))
	assert.False(t, cmd.Matches(component), "config has no message components")
	assert.Equal(t, "config", cmd.GetName())
}

func TestGetDescriptions_OneEntryPerSubCommandWithDiscordID(t *testing.T) {
	cmd := newFixture(t).cmd

	descriptions := cmd.GetDescriptions(discordgo.French)

	require.Len(t, descriptions, 5)
	for i, sub := range []string{"get", "almanax", "rss", "server", "twitter"} {
		assert.Equal(t, "/config "+sub, descriptions[i].Name)
		assert.Equal(t, "</config "+sub+":config-discord-id>", descriptions[i].CommandID)
		assert.Equal(t, testutils.Translations(t, discordgo.French, "config.help.tutorial."+sub)[0],
			descriptions[i].TutorialURL)
		assert.NotEmpty(t, descriptions[i].Description)
		assert.NotContains(t, descriptions[i].Description, "{{", "placeholders are resolved")
	}
	assert.Equal(t, testutils.Translations(t, discordgo.French, "config.help.detailed.get")[0],
		descriptions[0].Description)
	assert.Contains(t, descriptions[1].Description, "*"+i18n.DefaultLocale.String()+"*",
		"almanax help advertises the default locale")
}

func TestGetDescriptions_Localized(t *testing.T) {
	cmd := newFixture(t).cmd

	fr := cmd.GetDescriptions(discordgo.French)
	en := cmd.GetDescriptions(discordgo.EnglishGB)

	assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "config.help.detailed.rss")[0],
		en[2].Description)
	assert.NotEqual(t, fr[2].Description, en[2].Description)
}

// ---- checks ----

type labelledCheck struct {
	name       string
	optionName string
	i18nKey    string
	setFound   func(f *fixture, count int)
	calls      func(f *fixture) []findCall
	ctxKey     constants.ContextKey
	check      func(f *fixture) func(context.Context, *discordgo.Session,
		*discordgo.InteractionCreate, func(context.Context))
}

func labelledChecks() []labelledCheck {
	return []labelledCheck{
		{
			name:       "feed type",
			optionName: contract.ConfigFeedTypeOptionName,
			i18nKey:    "checks.feed",
			setFound: func(f *fixture, count int) {
				f.feeds.found = nil
				for i := range count {
					f.feeds.found = append(f.feeds.found, feedType(string(rune('a'+i)), "Flux "+string(rune('A'+i))))
				}
			},
			calls:  func(f *fixture) []findCall { return f.feeds.calls },
			ctxKey: constants.ContextKeyFeed,
			check: func(f *fixture) func(context.Context, *discordgo.Session,
				*discordgo.InteractionCreate, func(context.Context)) {
				return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
					next func(context.Context)) {
					f.cmd.checkFeedType(ctx, s, i, next)
				}
			},
		},
		{
			name:       "twitter account",
			optionName: contract.ConfigTwitterAccountOptionName,
			i18nKey:    "checks.twitterAccount",
			setFound: func(f *fixture, count int) {
				f.twitters.found = nil
				for i := range count {
					f.twitters.found = append(f.twitters.found, entities.TwitterAccount{
						ID: string(rune('a' + i)), Name: "Account " + string(rune('A'+i)),
					})
				}
			},
			calls:  func(f *fixture) []findCall { return f.twitters.calls },
			ctxKey: constants.ContextKeyTwitter,
			check: func(f *fixture) func(context.Context, *discordgo.Session,
				*discordgo.InteractionCreate, func(context.Context)) {
				return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
					next func(context.Context)) {
					f.cmd.checkTwitterAccount(ctx, s, i, next)
				}
			},
		},
	}
}

func TestLabelledChecks_ExactlyOneMatchIsPutInContext(t *testing.T) {
	for _, tc := range labelledChecks() {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			tc.setFound(f, 1)
			s, fake := testutils.NewDiscordSession(t)
			next := &nextRecorder{}

			tc.check(f)(context.Background(), s,
				configCommand("sub", stringOption(tc.optionName, "  quer")), next.next)

			require.True(t, next.called)
			assert.NotNil(t, next.ctx.Value(tc.ctxKey))
			switch value := next.ctx.Value(tc.ctxKey).(type) {
			case entities.FeedType:
				assert.Equal(t, "a", value.ID)
			case entities.TwitterAccount:
				assert.Equal(t, "a", value.ID)
			default:
				t.Fatalf("unexpected context value %T", value)
			}
			assert.Equal(t, []findCall{{"  quer", discordgo.French, constants.MaxChoices}}, tc.calls(f),
				"raw option value, caller locale and discord choice limit are forwarded")
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestLabelledChecks_NoOrManyMatchesAnswerWithLocalizedErrorAndStop(t *testing.T) {
	for _, tc := range labelledChecks() {
		for _, count := range []int{0, 2} {
			t.Run(fmt.Sprintf("%s/%d matches", tc.name, count), func(t *testing.T) {
				f := newFixture(t)
				tc.setFound(f, count)
				s, fake := testutils.NewDiscordSession(t)
				next := &nextRecorder{}

				tc.check(f)(context.Background(), s,
					configCommand("sub", stringOption(tc.optionName, "zzz")), next.next)

				assert.False(t, next.called)
				assertEditedWebhook(t, fake)
				content := editContent(t, fake)
				assert.Contains(t, content, "'*zzz*'")
				if count == 0 {
					assert.Contains(t, content, "ne correspond à aucun")
				} else {
					assert.Contains(t, content, "correspond à plusieurs")
					assert.Contains(t, content, "(", "the ambiguous labels are listed")
				}
			})
		}
	}
}

func TestLabelledChecks_AmbiguousMessageListsLabels(t *testing.T) {
	f := newFixture(t)
	f.feeds.found = []entities.FeedType{feedType("news", "Actualités"), feedType("devblog", "Devblog")}
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.checkFeedType(context.Background(), s,
		configCommand("rss", stringOption(contract.ConfigFeedTypeOptionName, "a")), (&nextRecorder{}).next)

	assert.Contains(t, editContent(t, fake), "(Actualités, Devblog)")
}

func TestLabelledChecks_OptionAbsentPassesThroughUnchanged(t *testing.T) {
	for _, tc := range labelledChecks() {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)
			next := &nextRecorder{}

			tc.check(f)(context.Background(), s,
				configCommand("sub", enabledOption(true)), next.next)

			require.True(t, next.called)
			assert.Nil(t, next.ctx.Value(tc.ctxKey))
			assert.Empty(t, tc.calls(f))
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestLabelledChecks_EditFailureIsSwallowed(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 400, Body: `{"code":0}`})
	next := &nextRecorder{}

	assert.NotPanics(t, func() {
		f.cmd.checkTwitterAccount(context.Background(), s,
			configCommand("twitter", stringOption(contract.ConfigTwitterAccountOptionName, "x")), next.next)
	})
	assert.False(t, next.called)
	assert.Len(t, fake.Requests(), 1)
}

func TestCheckChannelID_ExplicitChannelOptionWins(t *testing.T) {
	f := newFixture(t)
	s, _ := testutils.NewDiscordSession(t)
	addGuild(t, s, 0, textChannel(otherChannel))
	next := &nextRecorder{}

	f.cmd.checkChannelID(context.Background(), s,
		configCommand(contract.ConfigAlmanaxSubCommandName,
			enabledOption(true), otherChannelOption()), next.next)

	require.True(t, next.called)
	assert.Equal(t, otherChannel, next.ctx.Value(constants.ContextKeyChannel))
}

func TestCheckChannelID_WebhookSubCommandsDefaultToCurrentChannel(t *testing.T) {
	for _, sub := range []string{
		contract.ConfigAlmanaxSubCommandName, contract.ConfigRSSSubCommandName,
		contract.ConfigTwitterSubCommandName,
	} {
		t.Run(sub, func(t *testing.T) {
			f := newFixture(t)
			s, _ := testutils.NewDiscordSession(t)
			next := &nextRecorder{}

			f.cmd.checkChannelID(context.Background(), s,
				configCommand(sub, enabledOption(false)), next.next)

			require.True(t, next.called)
			assert.Equal(t, currentChannel, next.ctx.Value(constants.ContextKeyChannel))
		})
	}
}

func TestCheckChannelID_ServerSubCommandWithoutChannelTargetsWholeGuild(t *testing.T) {
	f := newFixture(t)
	s, _ := testutils.NewDiscordSession(t)
	next := &nextRecorder{}

	f.cmd.checkChannelID(context.Background(), s,
		configCommand(contract.ConfigServerSubCommandName,
			stringOption(contract.ConfigServerOptionName, "Djaul")), next.next)

	require.True(t, next.called)
	assert.Nil(t, next.ctx.Value(constants.ContextKeyChannel),
		"no channel means the server is bound to the guild")
}

func TestCheckEnabled_PutsOptionValueInContext(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		f := newFixture(t)
		next := &nextRecorder{}

		f.cmd.checkEnabled(context.Background(), nil,
			configCommand(contract.ConfigAlmanaxSubCommandName,
				enabledOption(enabled)), next.next)

		require.True(t, next.called)
		assert.Equal(t, enabled, next.ctx.Value(constants.ContextKeyEnabled))
	}
}

func TestCheckEnabled_AbsentOptionPassesThrough(t *testing.T) {
	f := newFixture(t)
	next := &nextRecorder{}

	f.cmd.checkEnabled(context.Background(), nil,
		configCommand(contract.ConfigAlmanaxSubCommandName), next.next)

	require.True(t, next.called)
	assert.Nil(t, next.ctx.Value(constants.ContextKeyEnabled))
}

func channelCtx(channelID string) context.Context {
	return context.WithValue(context.Background(), constants.ContextKeyChannel, channelID)
}

func TestCheckFollowConstraints_TextChannelWithWebhookPermissionContinues(t *testing.T) {
	for name, permissions := range map[string]int64{
		"manage webhooks": discordgo.PermissionManageWebhooks | discordgo.PermissionViewChannel,
		"administrator":   discordgo.PermissionAdministrator,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)
			addGuild(t, s, permissions, textChannel(currentChannel))
			next := &nextRecorder{}
			ctx := channelCtx(currentChannel)

			f.cmd.checkFollowConstraints(ctx, s, configCommand("almanax"), next.next)

			require.True(t, next.called)
			assert.Equal(t, ctx, next.ctx, "context is forwarded untouched")
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestCheckFollowConstraints_NonTextChannelIsRejected(t *testing.T) {
	for name, channelType := range map[string]discordgo.ChannelType{
		"voice":        discordgo.ChannelTypeGuildVoice,
		"announcement": discordgo.ChannelTypeGuildNews,
		"forum":        discordgo.ChannelTypeGuildForum,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)
			addGuild(t, s, discordgo.PermissionAdministrator,
				&discordgo.Channel{ID: currentChannel, Type: channelType})
			next := &nextRecorder{}

			f.cmd.checkFollowConstraints(channelCtx(currentChannel), s, configCommand("almanax"), next.next)

			assert.False(t, next.called)
			assertEditedWebhook(t, fake)
			assert.Equal(t, testutils.Translations(t, discordgo.French, "checks.webhook.targetChannelType")[0],
				editContent(t, fake))
		})
	}
}

func TestCheckFollowConstraints_MissingWebhookPermissionIsRejected(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	addGuild(t, s, discordgo.PermissionSendMessages|discordgo.PermissionViewChannel,
		textChannel(currentChannel))
	next := &nextRecorder{}

	f.cmd.checkFollowConstraints(channelCtx(currentChannel), s, configCommand("almanax"), next.next)

	assert.False(t, next.called)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "checks.webhook.permission")[0],
		editContent(t, fake))
}

func TestCheckFollowConstraints_ChannelOverwriteDenyingWebhooksIsRejected(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	channel := textChannel(currentChannel)
	channel.PermissionOverwrites = []*discordgo.PermissionOverwrite{{
		ID: guildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: discordgo.PermissionManageWebhooks,
	}}
	addGuild(t, s, discordgo.PermissionManageWebhooks, channel)
	next := &nextRecorder{}

	f.cmd.checkFollowConstraints(channelCtx(currentChannel), s, configCommand("almanax"), next.next)

	assert.False(t, next.called)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "checks.webhook.permission")[0],
		editContent(t, fake))
}

func TestCheckFollowConstraints_PermissionLookupFailureCountsAsMissingPermission(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	require.NoError(t, s.State.GuildAdd(&discordgo.Guild{ID: guildID, OwnerID: "someone-else"}))
	require.NoError(t, s.State.ChannelAdd(&discordgo.Channel{
		ID: currentChannel, GuildID: guildID, Type: discordgo.ChannelTypeGuildText,
	}))
	// The bot member is not cached: discordgo falls back to REST, which fails.
	fake.Respond(testutils.DiscordResponse{PathContains: "/members/", Status: 404, Body: `{"code":10007}`})
	next := &nextRecorder{}

	f.cmd.checkFollowConstraints(channelCtx(currentChannel), s, configCommand("almanax"), next.next)

	assert.False(t, next.called)
	assert.Equal(t, "/api/v9/guilds/"+guildID+"/members/bot-id", fake.Requests()[0].Path)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "checks.webhook.permission")[0],
		editContent(t, fake))
}

func TestCheckFollowConstraints_UncachedChannelIsFetchedThroughREST(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	addGuild(t, s, discordgo.PermissionManageWebhooks|discordgo.PermissionViewChannel)
	fake.Respond(testutils.DiscordResponse{Method: "GET", PathContains: "/channels/" + currentChannel,
		Status: 200, Body: fmt.Sprintf(`{"id":%q,"guild_id":%q,"type":0}`, currentChannel, guildID)})
	next := &nextRecorder{}
	ctx := channelCtx(currentChannel)

	f.cmd.checkFollowConstraints(ctx, s, configCommand("almanax"), next.next)

	require.True(t, next.called, "a channel absent from the state is fetched through REST")
	assert.Equal(t, ctx, next.ctx, "context is forwarded untouched")
	require.NotEmpty(t, fake.Requests())
	assert.Equal(t, "/api/v9/channels/"+currentChannel, fake.Requests()[0].Path)
}

func TestCheckFollowConstraints_UncachedNonTextChannelIsRejected(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	addGuild(t, s, discordgo.PermissionAdministrator)
	fake.Respond(testutils.DiscordResponse{Method: "GET", PathContains: "/channels/" + currentChannel,
		Status: 200, Body: fmt.Sprintf(`{"id":%q,"guild_id":%q,"type":2}`, currentChannel, guildID)})
	next := &nextRecorder{}

	f.cmd.checkFollowConstraints(channelCtx(currentChannel), s, configCommand("almanax"), next.next)

	assert.False(t, next.called)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "checks.webhook.targetChannelType")[0],
		editContent(t, fake))
}

func TestCheckFollowConstraints_UnknownChannelPanics(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{Method: "GET", PathContains: "/channels/",
		Status: 404, Body: `{"message":"Unknown Channel","code":10003}`})
	next := &nextRecorder{}

	assert.Panics(t, func() {
		f.cmd.checkFollowConstraints(channelCtx("not-cached"), s, configCommand("almanax"), next.next)
	})
	assert.False(t, next.called)
}

func TestCheckFollowConstraints_MissingContextFallsBackToInteractionChannel(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	addGuild(t, s, discordgo.PermissionAdministrator, textChannel(currentChannel))
	next := &nextRecorder{}

	f.cmd.checkFollowConstraints(context.Background(), s, configCommand("almanax"), next.next)

	require.True(t, next.called)
	assert.Empty(t, fake.Requests(), "the interaction channel is the one being checked")
}

func TestCheckFollowConstraints_NoChannelAtAllPanics(t *testing.T) {
	f := newFixture(t)
	s, _ := testutils.NewDiscordSession(t)
	event := configCommand("almanax")
	event.ChannelID = ""
	next := &nextRecorder{}

	assert.Panics(t, func() {
		f.cmd.checkFollowConstraints(context.Background(), s, event, next.next)
	})
	assert.False(t, next.called)
}

// ---- autocomplete ----

type autocompletePayload struct {
	Type discordgo.InteractionResponseType `json:"type"`
	Data struct {
		Choices []*discordgo.ApplicationCommandOptionChoice `json:"choices"`
	} `json:"data"`
}

func autocompleteAnswer(t *testing.T, fake *testutils.FakeDiscord) autocompletePayload {
	t.Helper()
	req := fake.Last(t)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/api/v9/interactions/interaction-id/tok/callback", req.Path)
	var payload autocompletePayload
	req.Decode(t, &payload)
	assert.Equal(t, discordgo.InteractionApplicationCommandAutocompleteResult, payload.Type)
	return payload
}

func choiceNames(choices []*discordgo.ApplicationCommandOptionChoice) []string {
	names := make([]string, 0, len(choices))
	for _, choice := range choices {
		names = append(names, choice.Name)
		if choice.Value != choice.Name {
			names = append(names, fmt.Sprintf("value mismatch: %v", choice.Value))
		}
	}
	return names
}

func TestAutocomplete_ServersAreLocalizedAndSortedByLabel(t *testing.T) {
	f := newFixture(t)
	f.servers.found = []entities.Server{
		server("s1", "Tylézia", "Tylezia"),
		server("s2", "", "Draconiros"), // no FR label: falls back to the default locale
		server("s3", "Imagiro", "Imagiro EN"),
	}
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, configAutocomplete(contract.ConfigServerSubCommandName,
		focused(stringOption(contract.ConfigServerOptionName, "i"))))

	payload := autocompleteAnswer(t, fake)
	assert.Equal(t, []string{"Draconiros", "Imagiro", "Tylézia"}, choiceNames(payload.Data.Choices))
	assert.Equal(t, []findCall{{"i", discordgo.French, constants.MaxChoices}}, f.servers.calls)
}

func TestAutocomplete_FeedTypesAreSortedByLabel(t *testing.T) {
	f := newFixture(t)
	f.feeds.found = []entities.FeedType{feedType("b", "Zaap"), feedType("a", "Almanax")}
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, configAutocomplete(contract.ConfigRSSSubCommandName,
		enabledOption(true),
		focused(stringOption(contract.ConfigFeedTypeOptionName, ""))))

	assert.Equal(t, []string{"Almanax", "Zaap"}, choiceNames(autocompleteAnswer(t, fake).Data.Choices))
	assert.Equal(t, []findCall{{"", discordgo.French, constants.MaxChoices}}, f.feeds.calls,
		"an empty feed query still lists feed types")
}

func TestAutocomplete_TwitterAccountsRequireNonBlankQuery(t *testing.T) {
	f := newFixture(t)
	f.twitters.found = []entities.TwitterAccount{{ID: "2", Name: "KaellyBot"}, {ID: "1", Name: "DOFUSfr"}}
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, configAutocomplete(contract.ConfigTwitterSubCommandName,
		focused(stringOption(contract.ConfigTwitterAccountOptionName, "   "))))
	assert.Empty(t, autocompleteAnswer(t, fake).Data.Choices)
	assert.Empty(t, f.twitters.calls, "blank queries do not hit the service")

	f.cmd.Handle(s, configAutocomplete(contract.ConfigTwitterSubCommandName,
		focused(stringOption(contract.ConfigTwitterAccountOptionName, "o"))))
	assert.Equal(t, []string{"DOFUSfr", "KaellyBot"}, choiceNames(autocompleteAnswer(t, fake).Data.Choices))
	assert.Len(t, f.twitters.calls, 1)
}

func TestAutocomplete_OnlyFocusedOptionIsCompleted(t *testing.T) {
	f := newFixture(t)
	f.servers.found = []entities.Server{server("s1", "Djaul", "")}
	f.feeds.found = []entities.FeedType{feedType("a", "Almanax")}
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, configAutocomplete(contract.ConfigRSSSubCommandName,
		stringOption(contract.ConfigServerOptionName, "d"),
		focused(stringOption(contract.ConfigFeedTypeOptionName, "a"))))

	assert.Equal(t, []string{"Almanax"}, choiceNames(autocompleteAnswer(t, fake).Data.Choices))
	assert.Empty(t, f.servers.calls)
}

func TestAutocomplete_UnknownOrNoFocusedOptionAnswersNoChoices(t *testing.T) {
	cases := map[string]*discordgo.InteractionCreate{
		"unknown focused option": configAutocomplete(contract.ConfigRSSSubCommandName,
			focused(stringOption("unknown", "a"))),
		"nothing focused": configAutocomplete(contract.ConfigRSSSubCommandName,
			stringOption(contract.ConfigFeedTypeOptionName, "a")),
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)

			f.cmd.Handle(s, event)

			assert.Empty(t, autocompleteAnswer(t, fake).Data.Choices)
			assert.Empty(t, f.feeds.calls)
		})
	}
}

func TestAutocomplete_DiscordFailureIsSwallowed(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/callback", Status: 400, Body: `{}`})

	assert.NotPanics(t, func() {
		f.cmd.Handle(s, configAutocomplete(contract.ConfigServerSubCommandName,
			focused(stringOption(contract.ConfigServerOptionName, "a"))))
	})
}
