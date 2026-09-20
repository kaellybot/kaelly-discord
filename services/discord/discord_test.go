package discord

import (
	"errors"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

// fakeGuildService answers Exists from a fixed set and records the lookups.
type fakeGuildService struct {
	guilds.Service
	known map[string]bool
	err   error

	mu      sync.Mutex
	lookups []string
}

func (f *fakeGuildService) Exists(guildID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lookups = append(f.lookups, guildID)
	return f.known[guildID], f.err
}

// fakeEmojiService renders misc emojis as recognizable markers.
type fakeEmojiService struct {
	emojis.Service
}

func (fakeEmojiService) GetMiscStringEmoji(emojiID constants.EmojiMiscID) string {
	return "<emoji:" + string(emojiID) + ">"
}

func newService(t *testing.T, guildService *fakeGuildService,
	broker *testutils.FakeBroker, cmds ...commands.DiscordCommand) (*Impl, *testutils.FakeDiscord) {
	t.Helper()
	session, fake := testutils.NewDiscordSession(t)
	return &Impl{
		shardID:      3,
		session:      session,
		commands:     cmds,
		guildService: guildService,
		emojiService: fakeEmojiService{},
		broker:       broker,
	}, fake
}

func TestNew_ConfiguresShardAndIntentsWithoutConnecting(t *testing.T) {
	broker := &testutils.FakeBroker{}
	service, err := New("secret", 2, 5, nil, &fakeGuildService{}, fakeEmojiService{}, broker)
	require.NoError(t, err)

	require.NotNil(t, service.session.Identify.Shard)
	assert.Equal(t, [2]int{2, 5}, *service.session.Identify.Shard)
	assert.Equal(t, constants.GetIntents(), service.session.Identify.Intents)
	assert.Equal(t, "Bot secret", service.session.Identify.Token)
	assert.Equal(t, 2, service.shardID)
	assert.Same(t, broker, service.broker)
	assert.False(t, service.IsConnected(), "nothing is connected before Listen")
}

func TestIsConnected_FollowsSessionDataReady(t *testing.T) {
	service, _ := newService(t, &fakeGuildService{}, &testutils.FakeBroker{})
	assert.False(t, service.IsConnected())

	service.session.DataReady = true
	assert.True(t, service.IsConnected())

	assert.False(t, (&Impl{}).IsConnected(), "a service without session is never connected")
}

func TestShutdown_MarksSessionAsNotReady(t *testing.T) {
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{})
	service.session.DataReady = true

	service.Shutdown()

	assert.False(t, service.IsConnected())
	assert.Empty(t, fake.Requests())
}

func guildCreateEvent(guild *discordgo.Guild) *discordgo.GuildCreate {
	return &discordgo.GuildCreate{Guild: guild}
}

func TestGuildCreate_UnavailableGuildIsIgnored(t *testing.T) {
	guildService := &fakeGuildService{}
	broker := &testutils.FakeBroker{}
	service, fake := newService(t, guildService, broker)

	service.guildCreate(service.session, guildCreateEvent(&discordgo.Guild{
		ID: "g1", Unavailable: true, SystemChannelID: "c1", OwnerID: "owner",
	}))

	assert.Empty(t, guildService.lookups, "outages must not even hit the database")
	assert.Empty(t, broker.Calls())
	assert.Empty(t, fake.Requests())
}

func TestGuildCreate_KnownGuildIsIgnored(t *testing.T) {
	guildService := &fakeGuildService{known: map[string]bool{"g1": true}}
	broker := &testutils.FakeBroker{}
	service, fake := newService(t, guildService, broker)

	service.guildCreate(service.session, guildCreateEvent(&discordgo.Guild{
		ID: "g1", Name: "Guild", OwnerID: "owner",
	}))

	assert.Equal(t, []string{"g1"}, guildService.lookups)
	assert.Empty(t, broker.Calls(), "reconnections replay GUILD_CREATE for every known guild")
	assert.Empty(t, fake.Requests(), "known guilds are not welcomed again")
}

func TestGuildCreate_ExistenceFailureIsIgnored(t *testing.T) {
	guildService := &fakeGuildService{err: errors.New("db down")}
	broker := &testutils.FakeBroker{}
	service, fake := newService(t, guildService, broker)

	service.guildCreate(service.session, guildCreateEvent(&discordgo.Guild{
		ID: "g1", Name: "Guild", OwnerID: "owner",
	}))

	assert.Equal(t, []string{"g1"}, guildService.lookups)
	assert.Empty(t, broker.Calls())
	assert.Empty(t, fake.Requests())
}

func TestGuildCreate_NewGuildEmitsNewsAndWelcomes(t *testing.T) {
	broker := &testutils.FakeBroker{}
	service, fake := newService(t, &fakeGuildService{}, broker)
	guild := guildWithBot(everyone(welcomeMessagePermissions),
		textChannel("sys"))
	guild.Name = "My Guild"
	guild.MemberCount = 42
	guild.SystemChannelID = "sys"
	require.NoError(t, service.session.State.GuildAdd(guild))

	service.guildCreate(service.session, guildCreateEvent(guild))

	calls := broker.Calls()
	require.Len(t, calls, 1)
	call := calls[0]
	assert.Equal(t, "Emit", call.Method)
	assert.Equal(t, amqp.ExchangeNews, call.Exchange)
	assert.Equal(t, constants.GuildNewsRoutingKey, call.RoutingKey)
	assert.Equal(t, "g1", call.CorrelationID, "the guild ID is used as correlation ID")
	assert.Equal(t, amqp.RabbitMQMessage_NEWS_GUILD, call.Message.GetType())
	assert.Equal(t, amqp.Language_ANY, call.Message.GetLanguage())
	news := call.Message.GetNewsGuildMessage()
	require.NotNil(t, news)
	assert.Equal(t, "g1", news.GetId())
	assert.Equal(t, "My Guild", news.GetName())
	assert.Equal(t, int64(42), news.GetMemberCount())
	assert.Equal(t, amqp.NewsGuildMessage_CREATE, news.GetEvent())

	req := fake.Last(t)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/api/v9/channels/sys/messages", req.Path)
}

func TestGuildCreate_EmitFailureStillWelcomes(t *testing.T) {
	broker := &testutils.FakeBroker{Err: errors.New("amqp down")}
	service, fake := newService(t, &fakeGuildService{}, broker)
	guild := guildWithBot(everyone(welcomeMessagePermissions), textChannel("sys"))
	guild.SystemChannelID = "sys"
	require.NoError(t, service.session.State.GuildAdd(guild))

	service.guildCreate(service.session, guildCreateEvent(guild))

	assert.Len(t, broker.Calls(), 1)
	assert.Equal(t, "/api/v9/channels/sys/messages", fake.Last(t).Path)
}

func guildDeleteEvent(guild *discordgo.Guild, unavailable bool) *discordgo.GuildDelete {
	return &discordgo.GuildDelete{
		Guild:        &discordgo.Guild{ID: guild.ID, Unavailable: unavailable},
		BeforeDelete: guild,
	}
}

func TestGuildDelete_WithoutBeforeDeleteIsIgnored(t *testing.T) {
	// discordgo only fills BeforeDelete when the guild was in state; the shard
	// used to die on the nil dereference here.
	guildService := &fakeGuildService{known: map[string]bool{"g1": true}}
	broker := &testutils.FakeBroker{}
	service, fake := newService(t, guildService, broker)

	require.NotPanics(t, func() {
		service.guildDelete(service.session, &discordgo.GuildDelete{
			Guild: &discordgo.Guild{ID: "g1"}, BeforeDelete: nil,
		})
	})

	assert.Empty(t, guildService.lookups, "nothing to announce without the cached guild")
	assert.Empty(t, broker.Calls())
	assert.Empty(t, fake.Requests())
}

func TestGuildDelete_KnownGuildEmitsDeleteNews(t *testing.T) {
	guildService := &fakeGuildService{known: map[string]bool{"g1": true}}
	broker := &testutils.FakeBroker{}
	service, fake := newService(t, guildService, broker)

	service.guildDelete(service.session, guildDeleteEvent(&discordgo.Guild{
		ID: "g1", Name: "Leaving", MemberCount: 7,
	}, false))

	assert.Equal(t, []string{"g1"}, guildService.lookups)
	calls := broker.Calls()
	require.Len(t, calls, 1)
	call := calls[0]
	assert.Equal(t, "Emit", call.Method)
	assert.Equal(t, amqp.ExchangeNews, call.Exchange)
	assert.Equal(t, constants.GuildNewsRoutingKey, call.RoutingKey)
	assert.Equal(t, "g1", call.CorrelationID)
	news := call.Message.GetNewsGuildMessage()
	require.NotNil(t, news)
	assert.Equal(t, "g1", news.GetId())
	assert.Equal(t, "Leaving", news.GetName(), "name comes from the state snapshot before deletion")
	assert.Equal(t, int64(7), news.GetMemberCount())
	assert.Equal(t, amqp.NewsGuildMessage_DELETE, news.GetEvent())
	assert.Empty(t, fake.Requests(), "no message is sent to a guild we left")
}

func TestGuildDelete_IgnoredCases(t *testing.T) {
	cases := map[string]struct {
		guildService *fakeGuildService
		unavailable  bool
		lookups      []string
	}{
		"outage": {
			guildService: &fakeGuildService{known: map[string]bool{"g1": true}},
			unavailable:  true,
		},
		"unknown guild": {
			guildService: &fakeGuildService{},
			lookups:      []string{"g1"},
		},
		"existence failure": {
			guildService: &fakeGuildService{known: map[string]bool{"g1": true}, err: errors.New("db down")},
			lookups:      []string{"g1"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			broker := &testutils.FakeBroker{}
			service, fake := newService(t, tc.guildService, broker)

			service.guildDelete(service.session, guildDeleteEvent(&discordgo.Guild{ID: "g1"}, tc.unavailable))

			assert.Equal(t, tc.lookups, tc.guildService.lookups)
			assert.Empty(t, broker.Calls())
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestGuildDelete_EmitFailureIsSwallowed(t *testing.T) {
	broker := &testutils.FakeBroker{Err: errors.New("amqp down")}
	service, _ := newService(t, &fakeGuildService{known: map[string]bool{"g1": true}}, broker)

	assert.NotPanics(t, func() {
		service.guildDelete(service.session, guildDeleteEvent(&discordgo.Guild{ID: "g1"}, false))
	})
	assert.Len(t, broker.Calls(), 1, "the emit was attempted once, without retry")
}
