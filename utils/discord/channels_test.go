package discord

import (
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cachedSession returns a session whose state already knows the given channels,
// the way discordgo caches a guild's channels in production.
func cachedSession(t *testing.T, channels ...*discordgo.Channel) (*discordgo.Session, *testutils.FakeDiscord) {
	t.Helper()
	s, fake := testutils.NewDiscordSession(t)
	require.NoError(t, s.State.GuildAdd(&discordgo.Guild{ID: "g1"}))
	for _, channel := range channels {
		require.NoError(t, s.State.ChannelAdd(channel))
	}

	return s, fake
}

func textChannel(id string) *discordgo.Channel {
	return &discordgo.Channel{ID: id, GuildID: "g1", Type: discordgo.ChannelTypeGuildText}
}

func thread(id, parentID string, threadType discordgo.ChannelType) *discordgo.Channel {
	return &discordgo.Channel{ID: id, GuildID: "g1", ParentID: parentID, Type: threadType}
}

func TestResolveParentChannelID_ThreadsAnswerThroughTheirParent(t *testing.T) {
	threadTypes := map[string]discordgo.ChannelType{
		"public thread":  discordgo.ChannelTypeGuildPublicThread,
		"private thread": discordgo.ChannelTypeGuildPrivateThread,
		"news thread":    discordgo.ChannelTypeGuildNewsThread,
	}
	for name, threadType := range threadTypes {
		t.Run(name, func(t *testing.T) {
			s, fake := cachedSession(t, textChannel("c1"), thread("t1", "c1", threadType))

			assert.Equal(t, "c1", ResolveParentChannelID(s, "t1"))
			assert.Empty(t, fake.Requests(), "a cached channel costs no API call")
		})
	}
}

func TestResolveParentChannelID_OtherChannelsAnswerForThemselves(t *testing.T) {
	cases := map[string]*discordgo.Channel{
		"text":     textChannel("c1"),
		"voice":    {ID: "c1", GuildID: "g1", Type: discordgo.ChannelTypeGuildVoice},
		"forum":    {ID: "c1", GuildID: "g1", Type: discordgo.ChannelTypeGuildForum},
		"category": {ID: "c1", GuildID: "g1", Type: discordgo.ChannelTypeGuildCategory},
		// A thread is not expected without a parent, but it must not resolve to "".
		"orphan thread": thread("c1", "", discordgo.ChannelTypeGuildPublicThread),
	}
	for name, channel := range cases {
		t.Run(name, func(t *testing.T) {
			s, _ := cachedSession(t, channel)
			assert.Equal(t, "c1", ResolveParentChannelID(s, "c1"))
		})
	}
}

func TestResolveParentChannelID_UncachedThreadIsFetched(t *testing.T) {
	s, fake := cachedSession(t)
	fake.Respond(testutils.DiscordResponse{Method: http.MethodGet, PathContains: "/channels/t1",
		Status: http.StatusOK,
		Body:   `{"id": "t1", "guild_id": "g1", "parent_id": "c1", "type": 11}`})

	assert.Equal(t, "c1", ResolveParentChannelID(s, "t1"))
	require.Len(t, fake.Requests(), 1)
	assert.Equal(t, "/api/v9/channels/t1", fake.Last(t).Path)
}

func TestResolveParentChannelID_UnknownChannelStandsForItself(t *testing.T) {
	s, fake := cachedSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/channels/ghost", Status: http.StatusNotFound,
		Body: `{"code": 10003, "message": "Unknown Channel"}`})

	assert.Equal(t, "ghost", ResolveParentChannelID(s, "ghost"),
		"a lookup failure must not lose the channel we were given")
}

func TestResolveParentChannelID_EmptyChannelIsNotLookedUp(t *testing.T) {
	s, fake := cachedSession(t)

	assert.Empty(t, ResolveParentChannelID(s, ""))
	assert.Empty(t, fake.Requests())
}
