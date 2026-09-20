package job

import (
	"context"
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	guildID   = "guild-1"
	channelID = "chan-1"
	authorID  = "author-1"
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

type fakeBooks struct {
	books.Service
	found []entities.Job
	jobs  []entities.Job
	calls []findCall
}

func (f *fakeBooks) FindJobs(name string, locale discordgo.Locale, limit int) []entities.Job {
	f.calls = append(f.calls, findCall{name, locale, limit})
	return f.found
}

func (f *fakeBooks) GetJobs() []entities.Job { return f.jobs }

func (f *fakeBooks) GetJob(id string) (entities.Job, bool) {
	for _, job := range f.jobs {
		if job.ID == id {
			return job, true
		}
	}
	return entities.Job{}, false
}

type fakeServers struct {
	servers.Service
	found []entities.Server
	calls []findCall
}

func (f *fakeServers) FindServers(name string, locale discordgo.Locale, limit int) []entities.Server {
	f.calls = append(f.calls, findCall{name, locale, limit})
	return f.found
}

func (f *fakeServers) GetServer(id string) (entities.Server, bool) {
	for _, server := range f.found {
		if server.ID == id {
			return server, true
		}
	}
	return entities.Server{}, false
}

type guildServerCall struct{ guildID, channelID string }

type fakeGuilds struct {
	guilds.Service
	server entities.Server
	found  bool
	err    error
	calls  []guildServerCall
}

func (f *fakeGuilds) GetServer(guildID, channelID string) (entities.Server, bool, error) {
	f.calls = append(f.calls, guildServerCall{guildID, channelID})
	return f.server, f.found, f.err
}

type fakeEmojis struct {
	emojis.Service
}

func (fakeEmojis) GetMiscEmoji(emojiID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: string(emojiID)}
}

func (fakeEmojis) GetEntityEmoji(entityID string, _ constants.EmojiEntity) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: entityID}
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
	cmd     *Command
	books   *fakeBooks
	servers *fakeServers
	guilds  *fakeGuilds
	manager *fakeRequestManager
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	viper.Set(constants.JobID, "job-discord-id")
	t.Cleanup(func() { viper.Set(constants.JobID, nil) })
	f := &fixture{
		books:   &fakeBooks{},
		servers: &fakeServers{},
		guilds:  &fakeGuilds{},
		manager: &fakeRequestManager{},
	}
	f.cmd = New(f.books, f.guilds, f.servers, fakeEmojis{}, f.manager)
	return f
}

// ---- entities ----

func job(id, frLabel string) entities.Job {
	return entities.Job{ID: id, Labels: []entities.JobLabel{
		{JobID: id, Locale: amqp.Language_FR, Label: frLabel},
		{JobID: id, Locale: amqp.Language_EN, Label: frLabel + " (en)"},
	}}
}

func server(id, label string) entities.Server {
	return entities.Server{ID: id, Labels: []entities.ServerLabel{
		{ServerID: id, Locale: amqp.Language_FR, Label: label},
		{ServerID: id, Locale: amqp.Language_EN, Label: label},
	}}
}

// ---- interactions ----

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		GuildID: guildID, ChannelID: channelID,
		Locale: discordgo.French, Data: data,
		Member: &discordgo.Member{User: &discordgo.User{ID: authorID}},
	}}
}

func jobCommand(subCommand string,
	options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand, discordgo.ApplicationCommandInteractionData{
		Name: contract.JobSlashCommandName,
		Options: []*discordgo.ApplicationCommandInteractionDataOption{{
			Name: subCommand, Type: discordgo.ApplicationCommandOptionSubCommand, Options: options,
		}},
	})
}

func jobAutocomplete(subCommand string,
	options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	event := jobCommand(subCommand, options...)
	event.Type = discordgo.InteractionApplicationCommandAutocomplete
	return event
}

func component(customID string, values ...string) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: customID, Values: values})
}

func stringOption(name, value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: name, Type: discordgo.ApplicationCommandOptionString, Value: value,
	}
}

func levelOption(level int64) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: contract.JobLevelOptionName, Type: discordgo.ApplicationCommandOptionInteger, Value: float64(level),
	}
}

func focused(option *discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.ApplicationCommandInteractionDataOption {
	option.Focused = true
	return option
}

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
	req := fake.Last(t)
	assert.Equal(t, "PATCH", req.Method)
	assert.Equal(t, "/api/v9/webhooks/app/tok/messages/@original", req.Path)
	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	return *payload.Content
}

// ---- Matches / descriptions ----

func TestMatches(t *testing.T) {
	f := newFixture(t)
	userCommand := func(name, targetID string) *discordgo.InteractionCreate {
		return interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: name, TargetID: targetID})
	}
	cases := map[string]struct {
		event *discordgo.InteractionCreate
		want  bool
	}{
		"slash command":             {jobCommand("get"), true},
		"autocomplete":              {jobAutocomplete("get"), true},
		"user command":              {userCommand(contract.JobUserCommandName, "target"), true},
		"slash name with target":    {userCommand(contract.JobSlashCommandName, "target"), false},
		"align command":             {userCommand("align", ""), false},
		"book page button":          {component("/books/job/alchemist?server=djaul&page=2"), true},
		"book job select":           {component("/books/job?server=djaul"), true},
		"alignment book component":  {component("/books/align?server=djaul"), false},
		"help component":            {component("/help"), false},
		"page button with bad page": {component("/books/job/alchemist?server=djaul&page=x"), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, f.cmd.Matches(tc.event))
		})
	}
	assert.Equal(t, "job", f.cmd.GetName())
}

func TestGetDescriptions_GetAndSetSubCommands(t *testing.T) {
	f := newFixture(t)

	descriptions := f.cmd.GetDescriptions(discordgo.French)

	require.Len(t, descriptions, 2)
	for i, sub := range []string{"get", "set"} {
		assert.Equal(t, "/job "+sub, descriptions[i].Name)
		assert.Equal(t, testutils.Translations(t, discordgo.French, "job.help.detailed."+sub)[0],
			descriptions[i].Description)
		assert.Equal(t, testutils.Translations(t, discordgo.French, "job.help.tutorial."+sub)[0],
			descriptions[i].TutorialURL)
		assert.Equal(t, "</job "+sub+":job-discord-id>", descriptions[i].CommandID,
			"the mention carries the configured JOB_ID")
	}
}

// ---- checkJob ----

func TestCheckJob_ExactlyOneMatchIsPutInContext(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("alchemist", "Alchimiste")}
	s, fake := testutils.NewDiscordSession(t)
	next := &nextRecorder{}

	f.cmd.checkJob(context.Background(), s,
		jobCommand("get", stringOption(contract.JobJobOptionName, "alch")), next.next)

	require.True(t, next.called)
	assert.Equal(t, f.books.found[0], next.ctx.Value(constants.ContextKeyJob))
	assert.Equal(t, []findCall{{"alch", discordgo.French, constants.MaxChoices}}, f.books.calls)
	assert.Empty(t, fake.Requests())
}

func TestCheckJob_NoOrManyMatchesAnswerLocalizedError(t *testing.T) {
	cases := map[string]struct {
		found []entities.Job
		want  string
	}{
		"none": {nil, "Aie! '*xx*' ne correspond à aucun métier que je connaisse ! :pensive:"},
		"many": {[]entities.Job{job("a", "Alchimiste"), job("b", "Bijoutier")},
			"Outch! '*xx*' correspond à plusieurs métiers (Alchimiste, Bijoutier) ! :slight_smile:"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			f.books.found = tc.found
			s, fake := testutils.NewDiscordSession(t)
			next := &nextRecorder{}

			f.cmd.checkJob(context.Background(), s,
				jobCommand("get", stringOption(contract.JobJobOptionName, "xx")), next.next)

			assert.False(t, next.called)
			assert.Equal(t, tc.want, editContent(t, fake))
		})
	}
}

func TestCheckJob_AbsentOptionPassesThrough(t *testing.T) {
	f := newFixture(t)
	next := &nextRecorder{}

	f.cmd.checkJob(context.Background(), nil,
		jobCommand("get", stringOption(contract.JobServerOptionName, "djaul")), next.next)

	require.True(t, next.called)
	assert.Nil(t, next.ctx.Value(constants.ContextKeyJob))
	assert.Empty(t, f.books.calls)
}

// ---- checkLevel ----

func TestCheckLevel_BoundsAreInclusive(t *testing.T) {
	for _, level := range []int64{constants.JobMinLevel, 1, constants.JobMaxLevel} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)
			next := &nextRecorder{}

			f.cmd.checkLevel(context.Background(), s, jobCommand("set", levelOption(level)), next.next)

			require.True(t, next.called)
			assert.Equal(t, level, next.ctx.Value(constants.ContextKeyLevel))
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestCheckLevel_OutOfRangeAnswersConstraintError(t *testing.T) {
	for _, level := range []int64{-1, constants.JobMaxLevel + 1} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)
			next := &nextRecorder{}

			f.cmd.checkLevel(context.Background(), s, jobCommand("set", levelOption(level)), next.next)

			assert.False(t, next.called)
			assert.Equal(t, "Le niveau doit être compris entre 0 et 200 ! :grin:", editContent(t, fake))
		})
	}
}

func TestCheckLevel_AbsentOptionStopsSilently(t *testing.T) {
	// Unlike the other checks, a missing level neither continues nor answers:
	// the interaction is left pending. Discord marks the option as required.
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	next := &nextRecorder{}

	f.cmd.checkLevel(context.Background(), s,
		jobCommand("set", stringOption(contract.JobJobOptionName, "alch")), next.next)

	assert.False(t, next.called)
	assert.Empty(t, fake.Requests())
}

// ---- autocomplete ----

type autocompletePayload struct {
	Type discordgo.InteractionResponseType `json:"type"`
	Data struct {
		Choices []*discordgo.ApplicationCommandOptionChoice `json:"choices"`
	} `json:"data"`
}

func autocompleteChoices(t *testing.T, fake *testutils.FakeDiscord) []string {
	t.Helper()
	req := fake.Last(t)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/api/v9/interactions/interaction-id/tok/callback", req.Path)
	var payload autocompletePayload
	req.Decode(t, &payload)
	assert.Equal(t, discordgo.InteractionApplicationCommandAutocompleteResult, payload.Type)
	names := make([]string, 0)
	for _, choice := range payload.Data.Choices {
		assert.Equal(t, choice.Name, choice.Value)
		names = append(names, choice.Name)
	}
	return names
}

func TestAutocomplete_JobsKeepServiceOrderAndLocale(t *testing.T) {
	f := newFixture(t)
	f.books.found = []entities.Job{job("tailor", "Tailleur"), job("alchemist", "Alchimiste")}
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, jobAutocomplete("get", focused(stringOption(contract.JobJobOptionName, "a"))))

	assert.Equal(t, []string{"Tailleur", "Alchimiste"}, autocompleteChoices(t, fake),
		"unlike config, job choices are not re-sorted")
	assert.Equal(t, []findCall{{"a", discordgo.French, constants.MaxChoices}}, f.books.calls)
}

func TestAutocomplete_Servers(t *testing.T) {
	f := newFixture(t)
	f.servers.found = []entities.Server{server("djaul", "Djaul")}
	s, fake := testutils.NewDiscordSession(t)
	event := jobAutocomplete("set", stringOption(contract.JobJobOptionName, "alch"),
		focused(stringOption(contract.JobServerOptionName, "dj")))
	event.Locale = discordgo.EnglishGB

	f.cmd.Handle(s, event)

	assert.Equal(t, []string{"Djaul"}, autocompleteChoices(t, fake))
	assert.Equal(t, []findCall{{"dj", discordgo.EnglishGB, constants.MaxChoices}}, f.servers.calls)
	assert.Empty(t, f.books.calls, "only the focused option is completed")
}

func TestAutocomplete_UnknownOrNoFocusedOptionAnswersEmpty(t *testing.T) {
	for name, event := range map[string]*discordgo.InteractionCreate{
		"unknown": jobAutocomplete("get", focused(stringOption("unknown", "x"))),
		"none":    jobAutocomplete("get", stringOption(contract.JobJobOptionName, "x")),
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)

			f.cmd.Handle(s, event)

			assert.Empty(t, autocompleteChoices(t, fake))
			assert.Empty(t, f.books.calls)
			assert.Empty(t, f.servers.calls)
		})
	}
}

func TestAutocomplete_DiscordFailureIsSwallowed(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/callback", Status: 400, Body: `{}`})

	assert.NotPanics(t, func() {
		f.cmd.Handle(s, jobAutocomplete("get", focused(stringOption(contract.JobJobOptionName, "x"))))
	})
}

// ---- option extraction ----

func TestOptionExtraction_MissingValuesAreErrors(t *testing.T) {
	ctx := context.Background()
	withJob := context.WithValue(ctx, constants.ContextKeyJob, job("a", "A"))
	withServer := context.WithValue(withJob, constants.ContextKeyServer, server("s", "S"))

	_, _, err := getGetOptions(ctx)
	require.Error(t, err)
	_, _, err = getGetOptions(withJob)
	require.Error(t, err)
	gotJob, gotServer, err := getGetOptions(withServer)
	require.NoError(t, err)
	assert.Equal(t, "a", gotJob.ID)
	assert.Equal(t, "s", gotServer.ID)

	_, _, _, err = getSetOptions(ctx)
	require.Error(t, err)
	_, _, _, err = getSetOptions(withJob)
	require.Error(t, err)
	_, _, _, err = getSetOptions(withServer)
	require.Error(t, err, "level is mandatory")
	_, _, _, err = getSetOptions(context.WithValue(withServer, constants.ContextKeyLevel, 12))
	require.Error(t, err, "level must be an int64")

	_, err = getUserOptions(ctx)
	require.Error(t, err)
}
