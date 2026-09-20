package discord

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
)

const required = discordgo.PermissionViewChannel | discordgo.PermissionManageWebhooks

// withPermissions makes the fake Discord describe channel c1 in guild g1 where
// the bot's only role grants rolePerms and the channel applies the given overwrites.
func withPermissions(fake *testutils.FakeDiscord, rolePerms int64, overwrites string) {
	fake.Respond(testutils.DiscordResponse{Method: http.MethodGet, PathContains: "/channels/c1",
		Status: http.StatusOK,
		Body:   fmt.Sprintf(`{"id": "c1", "guild_id": "g1", "permission_overwrites": [%s]}`, overwrites)})
	fake.Respond(testutils.DiscordResponse{Method: http.MethodGet, PathContains: "/guilds/g1",
		Status: http.StatusOK,
		Body: fmt.Sprintf(`{"id": "g1", "owner_id": "owner", "roles": [
			{"id": "g1", "permissions": "0"}, {"id": "bot-role", "permissions": "%d"}]}`, rolePerms)})
	fake.Respond(testutils.DiscordResponse{Method: http.MethodGet, PathContains: "/guilds/g1/members/bot-id",
		Status: http.StatusOK, Body: `{"user": {"id": "bot-id"}, "roles": ["bot-role"]}`})
}

func TestHasPermissions(t *testing.T) {
	cases := []struct {
		name       string
		rolePerms  int64
		overwrites string
		expected   bool
	}{
		{"all required bits granted", required | discordgo.PermissionSendMessages, "", true},
		{"only part of the required bits", discordgo.PermissionViewChannel, "", false},
		{"nothing granted", 0, "", false},
		{"administrator bypasses missing bits", discordgo.PermissionAdministrator, "", true},
		{"channel role overwrite denies a required bit", required,
			fmt.Sprintf(`{"id": "bot-role", "type": 0, "allow": "0", "deny": "%d"}`,
				discordgo.PermissionManageWebhooks), false},
		{"channel member overwrite grants the missing bit", discordgo.PermissionViewChannel,
			fmt.Sprintf(`{"id": "bot-id", "type": 1, "allow": "%d", "deny": "0"}`,
				discordgo.PermissionManageWebhooks), true},
		{"administrator survives channel denies", discordgo.PermissionAdministrator,
			fmt.Sprintf(`{"id": "g1", "type": 0, "allow": "0", "deny": "%d"}`, required), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, fake := testutils.NewDiscordSession(t)
			withPermissions(fake, tc.rolePerms, tc.overwrites)

			assert.Equal(t, tc.expected, HasPermissions(s, "c1", required))
		})
	}
}

func TestHasPermissions_ZeroRequirementIsAlwaysMet(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	withPermissions(fake, 0, "")

	assert.True(t, HasPermissions(s, "c1", 0))
}

func TestHasPermissions_ChecksTheBotUser(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	withPermissions(fake, required, "")

	assert.True(t, HasPermissions(s, "c1", required))

	paths := make([]string, 0)
	for _, req := range fake.Requests() {
		paths = append(paths, req.Path)
	}
	assert.Contains(t, paths, "/api/v"+discordgo.APIVersion+"/guilds/g1/members/bot-id")
}

func TestHasPermissions_UnknownChannelIsFalse(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/channels/c1", Status: http.StatusNotFound,
		Body: `{"code": 10003, "message": "Unknown Channel"}`})

	assert.False(t, HasPermissions(s, "c1", 0), "lookup failures are not treated as granted")
	assert.Len(t, fake.Requests(), 2,
		"one lookup to resolve a possible parent, one to compute the permissions")
}

// A thread owns no permission overwrite: everything is inherited from the
// channel it lives in, so the parent is what gets computed.
func TestHasPermissions_ThreadIsComputedOnItsParent(t *testing.T) {
	cases := map[string]struct {
		overwrites string
		expected   bool
	}{
		"parent grants the missing bits": {
			fmt.Sprintf(`{"id": "bot-id", "type": 1, "allow": "%d", "deny": "0"}`, required), true},
		"parent denies a required bit": {
			fmt.Sprintf(`{"id": "bot-id", "type": 1, "allow": "0", "deny": "%d"}`,
				discordgo.PermissionManageWebhooks), false},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			s, fake := cachedSession(t, thread("t1", "c1", discordgo.ChannelTypeGuildPublicThread))
			// The role alone grants nothing: only the parent's overwrite decides.
			withPermissions(fake, discordgo.PermissionViewChannel, testCase.overwrites)

			assert.Equal(t, testCase.expected, HasPermissions(s, "t1", required))
			paths := make([]string, 0)
			for _, req := range fake.Requests() {
				paths = append(paths, req.Path)
			}
			assert.Contains(t, paths, "/api/v"+discordgo.APIVersion+"/channels/c1",
				"the parent is the channel being resolved")
		})
	}
}
