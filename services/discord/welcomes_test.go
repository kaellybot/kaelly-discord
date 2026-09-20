package discord

import (
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	guildID         = "g1"
	systemChannelID = "sys"
)

// everyone is the @everyone role (its ID is the guild ID) with the given permissions.
func everyone(permissions int64) *discordgo.Role {
	return &discordgo.Role{Permissions: permissions}
}

func textChannel(id string, overwrites ...*discordgo.PermissionOverwrite) *discordgo.Channel {
	return &discordgo.Channel{ID: id, GuildID: guildID, Type: discordgo.ChannelTypeGuildText,
		PermissionOverwrites: overwrites}
}

func typedChannel(id string, channelType discordgo.ChannelType) *discordgo.Channel {
	return &discordgo.Channel{ID: id, GuildID: guildID, Type: channelType}
}

// denyEmbeds removes EmbedLinks for @everyone in a channel: the bot can talk
// there but not send the welcome embed.
func denyEmbeds() *discordgo.PermissionOverwrite {
	return &discordgo.PermissionOverwrite{ID: guildID, Type: discordgo.PermissionOverwriteTypeRole,
		Deny: discordgo.PermissionEmbedLinks}
}

// guildWithBot builds a guild where the bot is a plain member, so its
// permissions come from @everyone and the channel overwrites.
func guildWithBot(everyoneRole *discordgo.Role, channels ...*discordgo.Channel) *discordgo.Guild {
	everyoneRole.ID = guildID
	return &discordgo.Guild{
		ID:              guildID,
		Name:            "Guild " + guildID,
		OwnerID:         "owner-1",
		PreferredLocale: string(discordgo.EnglishUS),
		Roles:           []*discordgo.Role{everyoneRole},
		Channels:        channels,
		Members: []*discordgo.Member{{
			GuildID: guildID,
			User:    &discordgo.User{ID: "bot-id"},
		}},
	}
}

type sentMessage struct {
	Embeds []*discordgo.MessageEmbed `json:"embeds"`
}

// welcomeTargets returns the channel IDs that received a message, in order.
func welcomeTargets(fake *testutils.FakeDiscord) []string {
	targets := make([]string, 0)
	for _, req := range fake.Requests() {
		if req.Method == http.MethodPost && strings.HasSuffix(req.Path, "/messages") {
			path := strings.TrimSuffix(strings.TrimPrefix(req.Path, "/api/v9/channels/"), "/messages")
			targets = append(targets, path)
		}
	}
	return targets
}

func welcomeEmbed(t *testing.T, fake *testutils.FakeDiscord) *discordgo.MessageEmbed {
	t.Helper()
	req := fake.Last(t)
	require.True(t, strings.HasSuffix(req.Path, "/messages"), "last request is %s %s", req.Method, req.Path)
	var message sentMessage
	req.Decode(t, &message)
	require.Len(t, message.Embeds, 1)
	return message.Embeds[0]
}

func welcome(t *testing.T, guild *discordgo.Guild) *testutils.FakeDiscord {
	t.Helper()
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{})
	require.NoError(t, service.session.State.GuildAdd(guild))
	service.welcomeGuild(guild)
	return fake
}

func TestWelcomeGuild_ChannelSelection(t *testing.T) {
	cases := map[string]struct {
		guild  *discordgo.Guild
		target string
	}{
		"system channel wins over other writable channels": {
			guild: withSystemChannel(guildWithBot(everyone(welcomeMessagePermissions),
				textChannel("general"), textChannel("sys"))),
			target: "sys",
		},
		"system channel without embed rights falls back to first writable text channel": {
			guild: withSystemChannel(guildWithBot(everyone(welcomeMessagePermissions),
				textChannel("sys", denyEmbeds()),
				textChannel("no-embeds", denyEmbeds()),
				textChannel("first-ok"),
				textChannel("second-ok"))),
			target: "first-ok",
		},
		"no system channel uses the first writable text channel": {
			guild: guildWithBot(everyone(welcomeMessagePermissions),
				textChannel("first-ok"), textChannel("second-ok")),
			target: "first-ok",
		},
		"non-text channels are skipped even when writable": {
			guild: guildWithBot(everyone(welcomeMessagePermissions),
				typedChannel("voice", discordgo.ChannelTypeGuildVoice),
				typedChannel("category", discordgo.ChannelTypeGuildCategory),
				typedChannel("news", discordgo.ChannelTypeGuildNews),
				typedChannel("forum", discordgo.ChannelTypeGuildForum),
				textChannel("text")),
			target: "text",
		},
		"send messages alone is not enough": {
			guild: guildWithBot(everyone(discordgo.PermissionSendMessages),
				textChannel("text", &discordgo.PermissionOverwrite{
					ID: "g1", Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionEmbedLinks,
				}),
				textChannel("never")),
			target: "text",
		},
		"administrator bypasses channel denies": {
			guild: withSystemChannel(guildWithBot(everyone(discordgo.PermissionAdministrator),
				textChannel("sys", &discordgo.PermissionOverwrite{
					ID: "g1", Type: discordgo.PermissionOverwriteTypeRole, Deny: welcomeMessagePermissions,
				}))),
			target: "sys",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fake := welcome(t, tc.guild)
			assert.Equal(t, []string{tc.target}, welcomeTargets(fake))
			assert.Len(t, fake.Requests(), 1, "a single message, no DM to the owner")
		})
	}
}

func withSystemChannel(guild *discordgo.Guild) *discordgo.Guild {
	guild.SystemChannelID = systemChannelID
	return guild
}

func TestWelcomeGuild_EmbedIsLocalizedWithGuildPreferredLocale(t *testing.T) {
	viper.Set(constants.HelpID, "4242")
	t.Cleanup(func() { viper.Set(constants.HelpID, nil) })

	guild := withSystemChannel(guildWithBot(everyone(welcomeMessagePermissions),
		textChannel("sys")))
	guild.Name = "Les Crâ"
	guild.PreferredLocale = string(discordgo.French)

	embed := welcomeEmbed(t, welcome(t, guild))

	assert.Contains(t, embed.Description, "Salut et merci pour l'invitation")
	assert.Contains(t, embed.Description, "<@owner-1>", "the owner is mentioned")
	assert.Contains(t, embed.Description, "**Les Crâ**")
	assert.Contains(t, embed.Description, ":4242>", "the help command mention uses the configured ID")
	assert.Contains(t, embed.Description, "<emoji:"+string(constants.EmojiIDGame)+">")
	assert.Contains(t, embed.Description, "<emoji:"+string(constants.EmojiIDCalendar)+">")
	assert.Equal(t, constants.Color, embed.Color)
	require.NotNil(t, embed.Footer)
	assert.Contains(t, embed.Footer.Text, "Commandes Slash disponibles")
}

func TestWelcomeGuild_EnglishUSGuildGetsEnglish(t *testing.T) {
	guild := withSystemChannel(guildWithBot(everyone(welcomeMessagePermissions),
		textChannel("sys")))

	embed := welcomeEmbed(t, welcome(t, guild))

	assert.Contains(t, embed.Description, "Hello and thanks for the invitation")
}

func TestWelcomeGuild_ChannelSendFailureDoesNotFallBack(t *testing.T) {
	guild := guildWithBot(everyone(welcomeMessagePermissions),
		textChannel("first"), textChannel("second"))
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{})
	require.NoError(t, service.session.State.GuildAdd(guild))
	fake.Respond(testutils.DiscordResponse{PathContains: "/messages", Status: 403, Body: `{"message":"no"}`})

	service.welcomeGuild(guild)

	assert.Equal(t, []string{"first"}, welcomeTargets(fake),
		"a failed send is only logged: no other channel, no DM")
	assert.Len(t, fake.Requests(), 1)
}

// ownerOnlyGuild has only channels the bot cannot write embeds into.
func ownerOnlyGuild() *discordgo.Guild {
	return withSystemChannel(guildWithBot(everyone(welcomeMessagePermissions),
		textChannel("sys", denyEmbeds()),
		textChannel("general", denyEmbeds()),
		typedChannel("voice", discordgo.ChannelTypeGuildVoice)))
}

func TestWelcomeGuild_NoWritableChannelDMsTheOwnerInTheirLocale(t *testing.T) {
	guild := ownerOnlyGuild()
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{})
	require.NoError(t, service.session.State.GuildAdd(guild))
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/users/@me/channels",
		Status: 200, Body: `{"id":"dm-1","type":1}`})
	fake.Respond(testutils.DiscordResponse{Method: "GET", PathContains: "/users/owner-1",
		Status: 200, Body: `{"id":"owner-1","locale":"fr"}`})

	service.welcomeGuild(guild)

	requests := fake.Requests()
	require.Len(t, requests, 3)
	assert.Equal(t, "POST", requests[0].Method)
	assert.Equal(t, "/api/v9/users/@me/channels", requests[0].Path)
	var dm struct {
		RecipientID string `json:"recipient_id"`
	}
	requests[0].Decode(t, &dm)
	assert.Equal(t, "owner-1", dm.RecipientID)
	assert.Equal(t, "GET", requests[1].Method)
	assert.Equal(t, "/api/v9/users/owner-1", requests[1].Path)
	assert.Equal(t, []string{"dm-1"}, welcomeTargets(fake))

	embed := welcomeEmbed(t, fake)
	assert.Contains(t, embed.Description, "Salut et merci pour l'invitation",
		"the owner's locale wins over the guild's en-US")
	assert.Contains(t, embed.Description, "**Guild g1**")
}

func TestWelcomeGuild_OwnerLookupFailureFallsBackToDefaultLocale(t *testing.T) {
	guild := ownerOnlyGuild()
	guild.PreferredLocale = string(discordgo.French)
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{})
	require.NoError(t, service.session.State.GuildAdd(guild))
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/users/@me/channels",
		Status: 200, Body: `{"id":"dm-1","type":1}`})
	fake.Respond(testutils.DiscordResponse{Method: "GET", PathContains: "/users/owner-1",
		Status: 404, Body: `{"message":"Unknown User"}`})

	service.welcomeGuild(guild)

	assert.Equal(t, []string{"dm-1"}, welcomeTargets(fake))
	embed := welcomeEmbed(t, fake)
	assert.Contains(t, embed.Description, "Hello and thanks for the invitation",
		"default locale (en-GB), not the guild's preferred one")
}

func TestWelcomeGuild_DMChannelFailureStopsSilently(t *testing.T) {
	guild := ownerOnlyGuild()
	service, fake := newService(t, &fakeGuildService{}, &testutils.FakeBroker{})
	require.NoError(t, service.session.State.GuildAdd(guild))
	fake.Respond(testutils.DiscordResponse{Method: "POST", PathContains: "/users/@me/channels",
		Status: 403, Body: `{"message":"Cannot send messages to this user"}`})

	service.welcomeGuild(guild)

	requests := fake.Requests()
	require.Len(t, requests, 1, "no user lookup nor message once the DM channel cannot be opened")
	assert.Equal(t, "/api/v9/users/@me/channels", requests[0].Path)
}

func TestWelcomeGuild_NoChannelAndNoOwnerSendsNothing(t *testing.T) {
	guild := ownerOnlyGuild()
	guild.OwnerID = ""

	fake := welcome(t, guild)

	assert.Empty(t, fake.Requests())
}
