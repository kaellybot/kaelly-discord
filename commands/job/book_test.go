package job

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
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const membersBody = `[
	{"user":{"id":"u1","username":"alice"},"nick":"Ally"},
	{"user":{"id":"u2","username":"bob"}}
]`

func withMembers(fake *testutils.FakeDiscord) {
	fake.Respond(testutils.DiscordResponse{Method: "GET", PathContains: "/guilds/" + guildID + "/members",
		Status: 200, Body: membersBody})
}

func singleRequest(t *testing.T, f *fixture) requestCall {
	t.Helper()
	require.Len(t, f.manager.calls, 1)
	call := f.manager.calls[0]
	assert.Equal(t, constants.JobRequestRoutingKey, call.routingKey)
	assert.Equal(t, authorID, call.message.UserID)
	assert.Equal(t, amqp.Language_FR, call.message.Language)
	return call
}

func expectedNicknames() map[string]any {
	return map[string]any{"u1": "Ally", "u2": "bob"}
}

// ---- get ----

func TestHandle_GetRequestsBookForGuildMembers(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	f.servers.found = []entities.Server{server("djaul", "Djaul")}
	s, fake := testutils.NewDiscordSession(t)
	withMembers(fake)

	require.NoError(t, f.cmd.Handle(s, jobCommand(contract.JobGetSubCommandName,
		stringOption(contract.JobJobOptionName, "alch"), stringOption(contract.JobServerOptionName, "dja"))))

	call := singleRequest(t, f)
	assert.Equal(t, amqp.RabbitMQMessage_JOB_GET_BOOK_REQUEST, call.message.Type)
	req := call.message.GetJobGetBookRequest()
	assert.Equal(t, "alchemist", req.GetJobId())
	assert.Equal(t, "djaul", req.GetServerId())
	assert.Zero(t, req.GetOffset())
	assert.Equal(t, int64(constants.MaxBookRowPerEmbed), req.GetSize())
	assert.ElementsMatch(t, []string{"u1", "u2"}, req.GetUserIds())
	require.Len(t, call.properties, 1)
	assert.Equal(t, expectedNicknames(), call.properties[0], "nickname wins over username")
	assert.Empty(t, f.guilds.calls, "explicit server skips the guild fallback")

	require.Len(t, fake.Requests(), 1)
	assert.Equal(t, "/api/v9/guilds/"+guildID+"/members", fake.Requests()[0].Path)
}

func TestHandle_GetFallsBackToGuildServer(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	f.guilds.server, f.guilds.found = server("imagiro", "Imagiro"), true
	s, fake := testutils.NewDiscordSession(t)
	withMembers(fake)

	require.NoError(t,
		f.cmd.Handle(s, jobCommand(contract.JobGetSubCommandName, stringOption(contract.JobJobOptionName, "alch"))))

	assert.Equal(t, []guildServerCall{{guildID, channelID}}, f.guilds.calls)
	assert.Equal(t, "imagiro", singleRequest(t, f).message.GetJobGetBookRequest().GetServerId())
}

func TestHandle_GetWithoutAnyServerAnswersRequiredError(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t,
		f.cmd.Handle(s, jobCommand(contract.JobGetSubCommandName, stringOption(contract.JobJobOptionName, "alch"))))

	assert.Empty(t, f.manager.calls)
	assert.Contains(t, editContent(t, fake), "Il semble qu'il n'y a pas de serveur")
}

func TestHandle_GetGuildLookupFailureIsReported(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	f.guilds.err = errors.New("db down")
	s, fake := testutils.NewDiscordSession(t)

	err := f.cmd.Handle(s, jobCommand(contract.JobGetSubCommandName,
		stringOption(contract.JobJobOptionName, "alch")))

	require.ErrorContains(t, err, "db down")
	assert.ErrorIs(t, err, commands.ErrUnavailable, "a database that refused is worth a retry")
	assert.Empty(t, f.manager.calls)
	require.Len(t, fake.Requests(), 1)
	assert.Equal(t, "/api/v9/channels/"+channelID, fake.Requests()[0].Path,
		"only the parent channel lookup: the router answers the caller, not the command")
}

func TestHandle_GetUnknownJobStopsBeforeServerLookup(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t,
		f.cmd.Handle(s, jobCommand(contract.JobGetSubCommandName, stringOption(contract.JobJobOptionName, "zz"))))

	assert.Empty(t, f.guilds.calls)
	assert.Empty(t, f.manager.calls)
	assert.Contains(t, editContent(t, fake), "'*zz*'")
}

func TestHandle_GetMembersFailureIsReported(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	f.servers.found = []entities.Server{server("djaul", "Djaul")}
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/members", Status: 403, Body: `{"code":50001}`})

	err := f.cmd.Handle(s, jobCommand(contract.JobGetSubCommandName,
		stringOption(contract.JobJobOptionName, "alch"), stringOption(contract.JobServerOptionName, "dja")))

	require.ErrorIs(t, err, commands.ErrUnavailable,
		"Discord refusing the member list is someone else failing, not our logic")
	assert.NotErrorIs(t, err, commands.ErrStaleInteraction)
	assert.Empty(t, f.manager.calls)
}

func TestHandle_RequestFailureIsReported(t *testing.T) {
	f := newFixture(t)
	f.manager.err = errors.New("broker down")
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	f.servers.found = []entities.Server{server("djaul", "Djaul")}
	f.guilds.server, f.guilds.found = server("djaul", "Djaul"), true
	s, fake := testutils.NewDiscordSession(t)
	withMembers(fake)

	for name, event := range map[string]*discordgo.InteractionCreate{
		"get": jobCommand(contract.JobGetSubCommandName,
			stringOption(contract.JobJobOptionName, "alch"), stringOption(contract.JobServerOptionName, "dja")),
		"set": jobCommand(contract.JobSetSubCommandName, stringOption(contract.JobJobOptionName, "alch"),
			levelOption(10), stringOption(contract.JobServerOptionName, "dja")),
		"page":   component("/books/job/alchemist?server=djaul&page=1"),
		"select": component("/books/job?server=djaul", "alchemist"),
		"user":   userCommand(),
	} {
		assert.ErrorContains(t, f.cmd.Handle(s, event), "broker down", name)
	}
}

// ---- set ----

func TestHandle_SetRequestsLevelUpdateForAuthor(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	f.servers.found = []entities.Server{server("djaul", "Djaul")}
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, f.cmd.Handle(s, jobCommand(contract.JobSetSubCommandName,
		stringOption(contract.JobJobOptionName, "alch"), levelOption(200),
		stringOption(contract.JobServerOptionName, "dja"))))

	call := singleRequest(t, f)
	assert.Equal(t, amqp.RabbitMQMessage_JOB_SET_REQUEST, call.message.Type)
	req := call.message.GetJobSetRequest()
	assert.Equal(t, authorID, req.GetUserId())
	assert.Equal(t, "alchemist", req.GetJobId())
	assert.Equal(t, "djaul", req.GetServerId())
	assert.Equal(t, int64(200), req.GetLevel())
	assert.Empty(t, call.properties)
	assert.Empty(t, fake.Requests(), "no member list is needed to set a level")
}

func TestHandle_SetInvalidLevelStopsBeforeServerLookup(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, f.cmd.Handle(s, jobCommand(contract.JobSetSubCommandName,
		stringOption(contract.JobJobOptionName, "alch"), levelOption(201))))

	assert.Empty(t, f.guilds.calls)
	assert.Empty(t, f.manager.calls)
	assert.Contains(t, editContent(t, fake), "entre 0 et 200")
}

func TestSetBookReply_SuccessUsesAnswerLanguage(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, f.cmd.setBookReply(context.Background(), s, jobCommand("set"),
		&amqp.RabbitMQMessage{Status: amqp.RabbitMQMessage_SUCCESS, Language: amqp.Language_EN}, nil))

	assert.Equal(t, "Craftsmen book updated!", editContent(t, fake))
}

func TestSetBookReply_FailureIsReported(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	err := f.cmd.setBookReply(context.Background(), s, jobCommand("set"),
		&amqp.RabbitMQMessage{Status: amqp.RabbitMQMessage_FAILED}, nil)

	require.ErrorIs(t, err, commands.ErrInvalidAnswerMessage)
	assert.ErrorIs(t, err, commands.ErrUnavailable, "a backend that failed is worth a retry")
	assert.Empty(t, fake.Requests())
}

// ---- message components ----

func TestHandle_PageButtonRequestsRequestedPage(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	withMembers(fake)

	require.NoError(t, f.cmd.Handle(s, component(contract.CraftJobBookPageCustomID("alchemist", "djaul", 2))))

	call := singleRequest(t, f)
	req := call.message.GetJobGetBookRequest()
	assert.Equal(t, "alchemist", req.GetJobId())
	assert.Equal(t, "djaul", req.GetServerId())
	assert.Equal(t, int64(2*constants.MaxBookRowPerEmbed), req.GetOffset())
	assert.ElementsMatch(t, []string{"u1", "u2"}, req.GetUserIds())
	assert.Equal(t, expectedNicknames(), call.properties[0])
}

func TestHandle_JobSelectRequestsFirstPageOfSelectedJob(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	withMembers(fake)

	require.NoError(t, f.cmd.Handle(s, component(contract.CraftJobBookSelectCustomID("djaul"), "tailor")))

	call := singleRequest(t, f)
	req := call.message.GetJobGetBookRequest()
	assert.Equal(t, "tailor", req.GetJobId())
	assert.Equal(t, "djaul", req.GetServerId())
	assert.Zero(t, req.GetOffset())
	assert.Equal(t, expectedNicknames(), call.properties[0])
}

func TestHandle_ComponentMembersFailureIsReported(t *testing.T) {
	for name, event := range map[string]*discordgo.InteractionCreate{
		"page":   component(contract.CraftJobBookPageCustomID("alchemist", "djaul", 2)),
		"select": component(contract.CraftJobBookSelectCustomID("djaul"), "tailor"),
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)
			fake.Respond(testutils.DiscordResponse{PathContains: "/members", Status: 403, Body: `{"code":50001}`})

			require.ErrorIs(t, f.cmd.Handle(s, event), commands.ErrUnavailable)
			assert.Empty(t, f.manager.calls)
		})
	}
}

func TestHandle_InvalidComponentsAreReportedAsStale(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	cases := map[string]struct {
		run     func() error
		wantErr error
	}{
		"custom ID matching no job pattern": {
			func() error { return f.cmd.Handle(s, component("/books/align?server=djaul")) },
			commands.ErrNoProvidedHandler},
		"job select without value": {
			func() error { return f.cmd.Handle(s, component(contract.CraftJobBookSelectCustomID("djaul"))) },
			commands.ErrInvalidInteraction},
		"job select with two values": {
			func() error {
				return f.cmd.Handle(s, component(contract.CraftJobBookSelectCustomID("djaul"), "a", "b"))
			},
			commands.ErrInvalidInteraction},
		"page custom ID without page": {
			func() error { return f.cmd.updateBookPage(s, component("/books/job?server=djaul")) },
			commands.ErrInvalidInteraction},
		"select custom ID carrying a page": {
			func() error { return f.cmd.updateJobBook(s, component("/books/job/a?server=djaul&page=1", "a")) },
			commands.ErrInvalidInteraction},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.run()

			require.ErrorIs(t, err, tc.wantErr)
			assert.ErrorIs(t, err, commands.ErrStaleInteraction,
				"a component we cannot read is a stale click, not a crash")
		})
	}
	assert.Empty(t, f.manager.calls)
	assert.Empty(t, fake.Requests(), "the router answers the caller, not the command")
}

func TestHandle_UnhandledInteractionTypesAreIgnored(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t,
		f.cmd.Handle(s, interaction(discordgo.InteractionModalSubmit, discordgo.ModalSubmitInteractionData{})))
	userAutocomplete := userCommand()
	userAutocomplete.Type = discordgo.InteractionApplicationCommandAutocomplete
	require.NoError(t, f.cmd.Handle(s, userAutocomplete))
	assert.Empty(t, f.manager.calls)
	assert.Empty(t, fake.Requests())
}

// ---- book answers ----

func bookAnswer(craftsmen ...*amqp.JobGetBookAnswer_Craftsman) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_BOOK_ANSWER,
		Status:   amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		JobGetBookAnswer: &amqp.JobGetBookAnswer{
			JobId: "alchemist", ServerId: "djaul", Page: 0, Pages: 1, Total: int64(len(craftsmen)),
			Craftsmen: craftsmen,
		},
	}
}

func TestGetBookReply_ListsKnownCraftsmenByLevel(t *testing.T) {
	f := newFixture(t)
	f.books.jobs = []entities.Job{job("alchemist", "Alchimiste"), job("tailor", "Tailleur")}
	f.servers.found = []entities.Server{server("djaul", "Djaul")}
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, f.cmd.getBookReply(context.Background(), s, jobCommand("get"), bookAnswer(
		&amqp.JobGetBookAnswer_Craftsman{UserId: "u2", Level: 50},
		&amqp.JobGetBookAnswer_Craftsman{UserId: "gone", Level: 200},
		&amqp.JobGetBookAnswer_Craftsman{UserId: "u1", Level: 120},
	), expectedNicknames()))

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "Alchimiste (en)s list", embed.Title, "answer language (EN) drives the rendering")
	assert.Equal(t, "Djaul", embed.Footer.Text)
	assert.NotContains(t, embed.Description, "gone", "craftsmen who left the guild are skipped")
	assert.Contains(t, embed.Description, "Lvl. 120: **Ally**\nLvl. 50: **bob**")

	raw := strings.Join(rawComponents(payload), "")
	assert.Contains(t, raw, contract.CraftJobBookSelectCustomID("djaul"))
	assert.Contains(t, raw, `"value":"tailor"`)
}

func rawComponents(payload testutils.WebhookEditPayload) []string {
	raws := make([]string, 0, len(payload.Components))
	for _, c := range payload.Components {
		raws = append(raws, string(c))
	}
	return raws
}

func TestGetBookReply_InvalidAnswerIsReported(t *testing.T) {
	failed := bookAnswer()
	failed.Status = amqp.RabbitMQMessage_FAILED
	wrongType := bookAnswer()
	wrongType.Type = amqp.RabbitMQMessage_JOB_GET_USER_ANSWER
	missing := bookAnswer()
	missing.JobGetBookAnswer = nil

	for name, message := range map[string]*amqp.RabbitMQMessage{
		"failed": failed, "wrong type": wrongType, "missing answer": missing,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)

			err := f.cmd.getBookReply(context.Background(), s, jobCommand("get"), message, expectedNicknames())

			require.ErrorIs(t, err, commands.ErrInvalidAnswerMessage)
			assert.ErrorIs(t, err, commands.ErrUnavailable, "a backend that failed is worth a retry")
			assert.Empty(t, fake.Requests())
		})
	}
}

// ---- user command ----

func userCommand() *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand, discordgo.ApplicationCommandInteractionData{
		Name:     contract.JobUserCommandName,
		TargetID: "target-1",
		Resolved: &discordgo.ApplicationCommandInteractionDataResolved{
			Users:   map[string]*discordgo.User{"target-1": {ID: "target-1", Username: "carol"}},
			Members: map[string]*discordgo.Member{"target-1": {Nick: "Caro"}},
		},
	})
}

func TestHandle_UserCommandRequestsTargetJobsOnGuildServer(t *testing.T) {
	f := newFixture(t)
	f.guilds.server, f.guilds.found = server("imagiro", "Imagiro"), true
	s, fake := testutils.NewDiscordSession(t)
	// Cached like discordgo caches a guild's channels: resolving the channel
	// the interaction came from then costs no API call.
	require.NoError(t, s.State.GuildAdd(&discordgo.Guild{ID: guildID}))
	require.NoError(t, s.State.ChannelAdd(&discordgo.Channel{
		ID: channelID, GuildID: guildID, Type: discordgo.ChannelTypeGuildText}))

	require.NoError(t, f.cmd.Handle(s, userCommand()))

	call := singleRequest(t, f)
	assert.Equal(t, amqp.RabbitMQMessage_JOB_GET_USER_REQUEST, call.message.Type)
	req := call.message.GetJobGetUserRequest()
	assert.Equal(t, "target-1", req.GetUserId())
	assert.Equal(t, "imagiro", req.GetServerId())
	require.Len(t, call.properties, 1)
	member, ok := call.properties[0][userProperty].(*discordgo.Member)
	require.True(t, ok)
	assert.Equal(t, "Caro", member.Nick)
	require.NotNil(t, member.User, "resolved user is attached to the member")
	assert.Equal(t, "carol", member.User.Username)
	assert.Empty(t, fake.Requests())
}

func TestHandle_UserCommandWithoutGuildServerAnswersRequiredError(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, f.cmd.Handle(s, userCommand()))

	assert.Empty(t, f.manager.calls)
	assert.Contains(t, editContent(t, fake), "Il semble qu'il n'y a pas de serveur")
}

func TestHandle_UserCommandWithUnresolvedTargetIsReportedAsStale(t *testing.T) {
	f := newFixture(t)
	f.guilds.server, f.guilds.found = server("imagiro", "Imagiro"), true
	s, _ := testutils.NewDiscordSession(t)
	event := userCommand()
	data, ok := event.Data.(discordgo.ApplicationCommandInteractionData)
	require.True(t, ok)
	delete(data.Resolved.Members, "target-1")

	err := f.cmd.Handle(s, event)

	require.ErrorIs(t, err, commands.ErrInvalidInteraction)
	assert.ErrorIs(t, err, commands.ErrStaleInteraction,
		"an interaction whose target we cannot read is not a crash")
	assert.Empty(t, f.manager.calls)
}

func userAnswer(status amqp.RabbitMQMessage_Status) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_USER_ANSWER,
		Status:   status,
		Language: amqp.Language_EN,
		JobGetUserAnswer: &amqp.JobGetUserAnswer{ServerId: "imagiro", Jobs: []*amqp.JobGetUserAnswer_JobExperience{
			{JobId: "tailor", Level: 20}, {JobId: "alchemist", Level: 100},
		}},
	}
}

func TestUserRespond_RendersMemberJobs(t *testing.T) {
	f := newFixture(t)
	f.books.jobs = []entities.Job{job("alchemist", "Alchimiste"), job("tailor", "Tailleur")}
	f.servers.found = []entities.Server{server("imagiro", "Imagiro")}
	s, fake := testutils.NewDiscordSession(t)
	member := &discordgo.Member{Nick: "Caro", User: &discordgo.User{ID: "target-1", Username: "carol"}}

	require.NoError(t, f.cmd.userRespond(context.Background(), s, userCommand(),
		userAnswer(amqp.RabbitMQMessage_SUCCESS), map[string]any{userProperty: member}))

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "Caro jobs", embed.Title)
	assert.Equal(t, "Alchimiste (en) Lvl. 100\nTailleur (en) Lvl. 20\n", embed.Description)
	assert.Equal(t, "Imagiro", embed.Footer.Text)
}

func TestUserRespond_InvalidAnswerOrPropertyIsReported(t *testing.T) {
	member := &discordgo.Member{User: &discordgo.User{ID: "target-1"}}
	cases := map[string]struct {
		message    *amqp.RabbitMQMessage
		properties map[string]any
		wantErr    error
	}{
		"failed status": {userAnswer(amqp.RabbitMQMessage_FAILED),
			map[string]any{userProperty: member}, commands.ErrInvalidAnswerMessage},
		"missing property": {userAnswer(amqp.RabbitMQMessage_SUCCESS),
			map[string]any{}, commands.ErrRequestPropertyNotFound},
		"wrong property": {userAnswer(amqp.RabbitMQMessage_SUCCESS),
			map[string]any{userProperty: *member}, commands.ErrRequestPropertyNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)

			err := f.cmd.userRespond(context.Background(), s, userCommand(), tc.message, tc.properties)

			require.ErrorIs(t, err, tc.wantErr)
			assert.NotErrorIs(t, err, commands.ErrStaleInteraction)
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestHandle_CallbacksWiredToMatchingResponders(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	f.servers.found = []entities.Server{server("djaul", "Djaul")}
	f.books.jobs = f.books.found
	f.guilds.server, f.guilds.found = server("djaul", "Djaul"), true
	s, fake := testutils.NewDiscordSession(t)
	withMembers(fake)

	require.NoError(t,
		f.cmd.Handle(s, jobCommand(contract.JobGetSubCommandName, stringOption(contract.JobJobOptionName, "alch"))))
	require.NoError(t, f.cmd.Handle(s, jobCommand(contract.JobSetSubCommandName,
		stringOption(contract.JobJobOptionName, "alch"), levelOption(3))))
	require.NoError(t, f.cmd.Handle(s, userCommand()))
	require.Len(t, f.manager.calls, 3)

	get, set, user := f.manager.calls[0], f.manager.calls[1], f.manager.calls[2]
	require.NoError(t, get.callback(context.Background(), s, jobCommand("get"), bookAnswer(), get.properties[0]))
	assert.Equal(t, "Alchimiste (en)s list", fake.LastEdit(t).Embeds[0].Title)

	require.NoError(t, set.callback(context.Background(), s, jobCommand("set"),
		&amqp.RabbitMQMessage{Status: amqp.RabbitMQMessage_SUCCESS, Language: amqp.Language_EN}, nil))
	assert.Equal(t, "Craftsmen book updated!", editContent(t, fake))

	require.NoError(t, user.callback(context.Background(), s, userCommand(),
		userAnswer(amqp.RabbitMQMessage_SUCCESS), user.properties[0]))
	assert.Equal(t, "Caro jobs", fake.LastEdit(t).Embeds[0].Title)
}
