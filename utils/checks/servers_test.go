package checks

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

type findCall struct {
	name   string
	locale discordgo.Locale
	limit  int
}

type fakeServerService struct {
	servers.Service
	found []entities.Server
	calls []findCall
}

func (f *fakeServerService) FindServers(name string, locale discordgo.Locale, limit int) []entities.Server {
	f.calls = append(f.calls, findCall{name, locale, limit})
	return f.found
}

type fakeGuildService struct {
	guilds.Service
	server entities.Server
	found  bool
	err    error
	calls  [][2]string
}

func (f *fakeGuildService) GetServer(guildID, channelID string) (entities.Server, bool, error) {
	f.calls = append(f.calls, [2]string{guildID, channelID})
	return f.server, f.found, f.err
}

func server(id, frLabel string) entities.Server {
	return entities.Server{ID: id, Labels: []entities.ServerLabel{
		{ServerID: id, Locale: amqp.Language_FR, Label: frLabel},
	}}
}

// nextSpy records whether the chain continued and which server was in context.
type nextSpy struct {
	called    bool
	server    entities.Server
	hasServer bool
}

func (n *nextSpy) next(ctx context.Context) {
	n.called = true
	n.server, n.hasServer = ctx.Value(constants.ContextKeyServer).(entities.Server)
}

func stringOption(name, value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: name, Type: discordgo.ApplicationCommandOptionString, Value: value,
	}
}

func subCommand(name string, options ...*discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: name, Type: discordgo.ApplicationCommandOptionSubCommand, Options: options,
	}
}

func command(options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i1", AppID: "app", Token: "tok", GuildID: "g1", ChannelID: "c1",
		Type: discordgo.InteractionApplicationCommand, Locale: discordgo.French,
		Data: discordgo.ApplicationCommandInteractionData{Name: "job", Options: options},
	}}
}

func run(t *testing.T, middleware middlewares.MiddlewareCommand,
	event *discordgo.InteractionCreate) (*nextSpy, *testutils.FakeDiscord) {
	t.Helper()
	s, fake := testutils.NewDiscordSession(t)
	spy := &nextSpy{}
	middleware(context.Background(), s, event, spy.next)
	return spy, fake
}

func editedContent(t *testing.T, fake *testutils.FakeDiscord) string {
	t.Helper()
	req := fake.Last(t)
	assert.Equal(t, http.MethodPatch, req.Method)
	assert.Contains(t, req.Path, "/webhooks/app/tok/messages/@original")
	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	return *payload.Content
}

func TestCheckServer_SingleMatchIsPutInContext(t *testing.T) {
	djaul := server("djaul", "Djaul")
	service := &fakeServerService{found: []entities.Server{djaul}}

	spy, fake := run(t, CheckServer("server", service),
		command(subCommand("get", stringOption("job", "alchi"), stringOption("server", "dja"))))

	assert.True(t, spy.called)
	require.True(t, spy.hasServer)
	assert.Equal(t, djaul, spy.server)
	assert.Equal(t, []findCall{{"dja", discordgo.French, constants.MaxChoices}}, service.calls)
	assert.Empty(t, fake.Requests(), "the command answers, not the check")
}

func TestCheckServer_NoMatchAnswersNotFound(t *testing.T) {
	spy, fake := run(t, CheckServer("server", &fakeServerService{}),
		command(subCommand("get", stringOption("server", "xyz"))))

	assert.False(t, spy.called)
	assert.Equal(t, "Aie! '*xyz*' ne correspond à aucun serveur que je connaisse ! :pensive:",
		editedContent(t, fake))
}

func TestCheckServer_SeveralMatchesListsLocalizedLabels(t *testing.T) {
	service := &fakeServerService{found: []entities.Server{server("djaul", "Djaul"), server("dakal", "Dakal")}}

	spy, fake := run(t, CheckServer("server", service), command(subCommand("get", stringOption("server", "d"))))

	assert.False(t, spy.called)
	assert.Equal(t, "Outch! '*d*' correspond à plusieurs serveurs (Djaul, Dakal)! "+
		"Réessaie en sélectionnant un serveur de la liste :slight_smile:", editedContent(t, fake))
}

func TestCheckServer_OptionAbsentContinuesWithoutServer(t *testing.T) {
	service := &fakeServerService{}

	spy, fake := run(t, CheckServer("server", service), command(subCommand("get", stringOption("job", "alchi"))))

	assert.True(t, spy.called)
	assert.False(t, spy.hasServer)
	assert.Empty(t, service.calls)
	assert.Empty(t, fake.Requests())
}

func TestCheckServer_OnlyLooksInsideSubCommands(t *testing.T) {
	// A top-level option with the right name is not a subcommand option: ignored.
	service := &fakeServerService{}

	spy, _ := run(t, CheckServer("server", service), command(stringOption("server", "djaul")))

	assert.True(t, spy.called)
	assert.False(t, spy.hasServer)
	assert.Empty(t, service.calls)
}

func TestCheckServer_DiscordFailureIsSwallowed(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: http.StatusInternalServerError,
		Body: `{"message": "oops"}`})
	spy := &nextSpy{}

	require.NotPanics(t, func() {
		CheckServer("server", &fakeServerService{})(context.Background(), s,
			command(subCommand("get", stringOption("server", "xyz"))), spy.next)
	})
	assert.False(t, spy.called)
	assert.Len(t, fake.Requests(), 1)
}

func TestCheckServerWithFallback_FilledOptionIsResolvedAtAnyDepth(t *testing.T) {
	cases := map[string]*discordgo.InteractionCreate{
		"top level":     command(stringOption("server", "dja")),
		"in subcommand": command(subCommand("get", stringOption("server", "dja"))),
		"in subcommand group": command(&discordgo.ApplicationCommandInteractionDataOption{
			Name: "channel", Type: discordgo.ApplicationCommandOptionSubCommandGroup,
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				subCommand("set", stringOption("job", "alchi"), stringOption("server", "dja")),
			},
		}),
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			djaul := server("djaul", "Djaul")
			service := &fakeServerService{found: []entities.Server{djaul}}
			guildService := &fakeGuildService{}

			spy, fake := run(t, CheckServerWithFallback("server", service, guildService), event)

			assert.True(t, spy.called)
			assert.Equal(t, djaul, spy.server)
			assert.Equal(t, []findCall{{"dja", discordgo.French, constants.MaxChoices}}, service.calls)
			assert.Empty(t, guildService.calls, "an explicit server never looks at the guild")
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestCheckServerWithFallback_FilledButUnknownDoesNotFallBack(t *testing.T) {
	guildService := &fakeGuildService{server: server("djaul", "Djaul"), found: true}

	spy, fake := run(t, CheckServerWithFallback("server", &fakeServerService{}, guildService),
		command(subCommand("get", stringOption("server", "xyz"))))

	assert.False(t, spy.called)
	assert.Empty(t, guildService.calls)
	assert.Equal(t, "Aie! '*xyz*' ne correspond à aucun serveur que je connaisse ! :pensive:",
		editedContent(t, fake))
}

func TestCheckServerWithFallback_SeveralMatchesAreRejected(t *testing.T) {
	service := &fakeServerService{found: []entities.Server{server("djaul", "Djaul"), server("dakal", "Dakal")}}

	spy, fake := run(t, CheckServerWithFallback("server", service, &fakeGuildService{}),
		command(stringOption("server", "d")))

	assert.False(t, spy.called)
	assert.Contains(t, editedContent(t, fake), "(Djaul, Dakal)")
}

func TestCheckServerWithFallback_UsesGuildServerWhenOptionAbsent(t *testing.T) {
	imagiro := server("imagiro", "Imagiro")
	service := &fakeServerService{}
	guildService := &fakeGuildService{server: imagiro, found: true}

	spy, fake := run(t, CheckServerWithFallback("server", service, guildService),
		command(subCommand("get", stringOption("job", "alchi"))))

	assert.True(t, spy.called)
	assert.Equal(t, imagiro, spy.server)
	assert.Equal(t, [][2]string{{"g1", "c1"}}, guildService.calls, "guild and channel are both considered")
	assert.Empty(t, service.calls)
	assert.Empty(t, fake.Requests())
}

func TestCheckServerWithFallback_NonStringOptionWithSameNameIsIgnored(t *testing.T) {
	guildService := &fakeGuildService{server: server("imagiro", "Imagiro"), found: true}
	service := &fakeServerService{}

	spy, _ := run(t, CheckServerWithFallback("server", service, guildService),
		command(&discordgo.ApplicationCommandInteractionDataOption{
			Name: "server", Type: discordgo.ApplicationCommandOptionInteger, Value: float64(3),
		}))

	assert.True(t, spy.called)
	assert.Equal(t, "imagiro", spy.server.ID)
	assert.Empty(t, service.calls)
}

func TestCheckServerWithFallback_NoGuildServerAsksToConfigureOne(t *testing.T) {
	spy, fake := run(t, CheckServerWithFallback("server", &fakeServerService{}, &fakeGuildService{}),
		command(subCommand("get")))

	assert.False(t, spy.called)
	content := editedContent(t, fake)
	assert.Contains(t, content, "pas de serveur DOFUS lié à ta guilde")
	assert.Contains(t, content, "/config server")
}

func TestCheckServerWithFallback_NoGuildServerDiscordFailureIsSwallowed(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: http.StatusNotFound,
		Body: `{"code": 10015, "message": "Unknown Webhook"}`})
	spy := &nextSpy{}

	require.NotPanics(t, func() {
		CheckServerWithFallback("server", &fakeServerService{}, &fakeGuildService{})(
			context.Background(), s, command(), spy.next)
	})
	assert.False(t, spy.called)
	assert.Len(t, fake.Requests(), 1)
}

func TestCheckServerWithFallback_GuildLookupFailurePanics(t *testing.T) {
	errDB := errors.New("database is down")
	s, fake := testutils.NewDiscordSession(t)
	spy := &nextSpy{}

	assert.PanicsWithValue(t, errDB, func() {
		CheckServerWithFallback("server", &fakeServerService{}, &fakeGuildService{err: errDB})(
			context.Background(), s, command(), spy.next)
	}, "the panic is left to the command's HandlePanic")
	assert.False(t, spy.called)
	assert.Empty(t, fake.Requests())
}
