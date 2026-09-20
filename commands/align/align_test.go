package align

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
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

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

// --- fixtures -------------------------------------------------------------

func bonta() entities.City {
	return entities.City{ID: "bonta", Type: constants.CityTypeLight, Color: 0x0000ff, Icon: "https://bonta.png",
		Labels: []entities.CityLabel{{Locale: amqp.Language_FR, Label: "Bonta"}, {Locale: amqp.Language_EN, Label: "Bonta"}}}
}

func brakmar() entities.City {
	return entities.City{ID: "brakmar", Type: constants.CityTypeDark, Color: 0xff0000,
		Labels: []entities.CityLabel{{Locale: amqp.Language_FR, Label: "Brâkmar"}}}
}

func spirit() entities.Order {
	return entities.Order{ID: "spirit", Labels: []entities.OrderLabel{{Locale: amqp.Language_FR, Label: "Esprit"}}}
}

func heart() entities.Order {
	return entities.Order{ID: "heart", Labels: []entities.OrderLabel{{Locale: amqp.Language_FR, Label: "Coeur"}}}
}

func djaul() entities.Server {
	return entities.Server{ID: "djaul", Icon: "https://djaul.png", Labels: []entities.ServerLabel{
		{Locale: amqp.Language_FR, Label: "Djaul"}, {Locale: amqp.Language_EN, Label: "Djaul EN"},
	}}
}

func hel() entities.Server {
	return entities.Server{ID: "hel", Labels: []entities.ServerLabel{{Locale: amqp.Language_FR, Label: "Hel Munster"}}}
}

type findCall struct {
	name   string
	locale discordgo.Locale
	limit  int
}

type fakeBooks struct {
	books.Service
	cities    []entities.City
	orders    []entities.Order
	findCalls []findCall
}

func (f *fakeBooks) FindCities(name string, locale discordgo.Locale, limit int) []entities.City {
	f.findCalls = append(f.findCalls, findCall{name, locale, limit})
	return f.cities
}

func (f *fakeBooks) FindOrders(name string, locale discordgo.Locale, limit int) []entities.Order {
	f.findCalls = append(f.findCalls, findCall{name, locale, limit})
	return f.orders
}

func (f *fakeBooks) GetCities() []entities.City  { return f.cities }
func (f *fakeBooks) GetOrders() []entities.Order { return f.orders }

func (f *fakeBooks) GetCity(id string) (entities.City, bool) {
	for _, city := range f.cities {
		if city.ID == id {
			return city, true
		}
	}
	return entities.City{}, false
}

func (f *fakeBooks) GetOrder(id string) (entities.Order, bool) {
	for _, order := range f.orders {
		if order.ID == id {
			return order, true
		}
	}
	return entities.Order{}, false
}

type fakeServers struct {
	servers.Service
	servers   []entities.Server
	findCalls []findCall
}

func (f *fakeServers) FindServers(name string, locale discordgo.Locale, limit int) []entities.Server {
	f.findCalls = append(f.findCalls, findCall{name, locale, limit})
	return f.servers
}

func (f *fakeServers) GetServer(id string) (entities.Server, bool) {
	for _, server := range f.servers {
		if server.ID == id {
			return server, true
		}
	}
	return entities.Server{}, false
}

type fakeGuilds struct {
	guilds.Service
	server entities.Server
	found  bool
	err    error
	calls  [][2]string
}

func (f *fakeGuilds) GetServer(guildID, channelID string) (entities.Server, bool, error) {
	f.calls = append(f.calls, [2]string{guildID, channelID})
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

func (f *fakeRequestManager) single(t *testing.T) requestCall {
	t.Helper()
	require.Len(t, f.calls, 1)
	assert.Equal(t, "requests.books", f.calls[0].routingKey)
	assert.Equal(t, "author-id", f.calls[0].message.UserID)
	assert.Equal(t, amqp.Language_FR, f.calls[0].message.Language)
	return f.calls[0]
}

type deps struct {
	books    *fakeBooks
	servers  *fakeServers
	guilds   *fakeGuilds
	requests *fakeRequestManager
}

func newDeps() *deps {
	return &deps{
		books:    &fakeBooks{},
		servers:  &fakeServers{},
		guilds:   &fakeGuilds{},
		requests: &fakeRequestManager{},
	}
}

func newCommand(t *testing.T, d *deps) *Command {
	t.Helper()
	viper.Set(constants.AlignID, "align-discord-id")
	t.Cleanup(func() { viper.Set(constants.AlignID, nil) })
	return New(d.books, d.guilds, d.servers, fakeEmojis{}, d.requests)
}

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		GuildID: "guild-id", ChannelID: "channel-id",
		Locale: discordgo.French, Data: data,
		Member: &discordgo.Member{User: &discordgo.User{ID: "author-id"}},
	}}
}

func slash(subCommand string, options ...*discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand, discordgo.ApplicationCommandInteractionData{
		Name: "align",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{{
			Name: subCommand, Type: discordgo.ApplicationCommandOptionSubCommand, Options: options,
		}},
	})
}

func userCommand(targetID string, resolved *discordgo.ApplicationCommandInteractionDataResolved,
) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand, discordgo.ApplicationCommandInteractionData{
		Name: "Alignments", CommandType: discordgo.UserApplicationCommand,
		TargetID: targetID, Resolved: resolved,
	})
}

func stringOption(name, value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: name, Type: discordgo.ApplicationCommandOptionString, Value: value,
	}
}

func levelOption(value int64) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: "level", Type: discordgo.ApplicationCommandOptionInteger, Value: float64(value),
	}
}

func component(customID string, values ...string) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: customID, Values: values})
}

const guildMembers = `[
	{"user":{"id":"u1","username":"alice"},"nick":"Ali"},
	{"user":{"id":"u2","username":"bob"}}
]`

func respondMembers(fake *testutils.FakeDiscord) {
	fake.Respond(testutils.DiscordResponse{Method: "GET", PathContains: "/guilds/guild-id/members",
		Status: 200, Body: guildMembers})
}

// --- command metadata -----------------------------------------------------

func TestGetName_IsAlign(t *testing.T) {
	assert.Equal(t, "align", newCommand(t, newDeps()).GetName())
}

func TestGetDescriptions_GetAndSet(t *testing.T) {
	descriptions := newCommand(t, newDeps()).GetDescriptions(discordgo.French)

	require.Len(t, descriptions, 2)
	for index, sub := range []string{"get", "set"} {
		assert.Equal(t, "/align "+sub, descriptions[index].Name)
		assert.Equal(t, "</align "+sub+":align-discord-id>", descriptions[index].CommandID)
		assert.Equal(t, testutils.Translations(t, discordgo.French, "align.help.detailed."+sub)[0],
			descriptions[index].Description)
		assert.Equal(t, testutils.Translations(t, discordgo.French, "align.help.tutorial."+sub)[0],
			descriptions[index].TutorialURL)
	}
}

func TestMatches(t *testing.T) {
	cmd := newCommand(t, newDeps())
	alignWithTarget := slash("get")
	data := alignWithTarget.ApplicationCommandData()
	data.TargetID = "u1"
	alignWithTarget.Data = data
	autocomplete := slash("get")
	autocomplete.Type = discordgo.InteractionApplicationCommandAutocomplete
	userNameWithoutTarget := userCommand("", nil)

	cases := []struct {
		name     string
		event    *discordgo.InteractionCreate
		expected bool
	}{
		{"slash command", slash("get"), true},
		{"autocomplete", autocomplete, true},
		{"slash name with a target", alignWithTarget, false},
		{"user command", userCommand("u1", nil), true},
		{"user name without a target", userNameWithoutTarget, false},
		{"other command", interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "job"}), false},
		{"page button", component("/books/align?city=_&order=_&server=djaul&page=1"), true},
		{"city select", component("/books/align?order=_&server=djaul"), true},
		{"order select", component("/books/align?city=_&server=djaul"), true},
		{"job book button", component("/books/job?job=_&server=djaul&page=1"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, cmd.Matches(tc.event))
		})
	}
}

// --- dispatch -------------------------------------------------------------

func TestHandle_GetRequestsBookWithGuildMembers(t *testing.T) {
	d := newDeps()
	d.books.cities = []entities.City{bonta()}
	d.books.orders = []entities.Order{spirit()}
	d.servers.servers = []entities.Server{djaul()}
	s, fake := testutils.NewDiscordSession(t)
	respondMembers(fake)

	newCommand(t, d).Handle(s, slash("get",
		stringOption("city", "bon"), stringOption("order", "esp"), stringOption("server", "dja")))

	req := fake.Last(t)
	assert.Equal(t, "GET", req.Method)
	assert.True(t, strings.HasSuffix(req.Path, "/guilds/guild-id/members"), req.Path)
	call := d.requests.single(t)
	assert.Equal(t, amqp.RabbitMQMessage_ALIGN_GET_BOOK_REQUEST, call.message.Type)
	request := call.message.GetAlignGetBookRequest()
	assert.Equal(t, "bonta", request.GetCityId())
	assert.Equal(t, "spirit", request.GetOrderId())
	assert.Equal(t, "djaul", request.GetServerId())
	assert.Equal(t, int64(0), request.GetOffset())
	assert.Equal(t, int64(constants.MaxBookRowPerEmbed), request.GetSize())
	assert.ElementsMatch(t, []string{"u1", "u2"}, request.GetUserIds())
	require.Len(t, call.properties, 1)
	assert.Equal(t, map[string]any{"u1": "Ali", "u2": "bob"}, call.properties[0],
		"nickname wins over username")
	assert.Empty(t, d.guilds.calls, "explicit server skips the guild fallback")
}

func TestHandle_GetWithoutFiltersUsesGuildServer(t *testing.T) {
	d := newDeps()
	d.guilds.server, d.guilds.found = hel(), true
	s, fake := testutils.NewDiscordSession(t)
	respondMembers(fake)

	newCommand(t, d).Handle(s, slash("get"))

	assert.Equal(t, [][2]string{{"guild-id", "channel-id"}}, d.guilds.calls)
	request := d.requests.single(t).message.GetAlignGetBookRequest()
	assert.Empty(t, request.GetCityId())
	assert.Empty(t, request.GetOrderId())
	assert.Equal(t, "hel", request.GetServerId())
}

func TestHandle_GetWithoutAnyServerAnswersServerRequired(t *testing.T) {
	d := newDeps()
	s, fake := testutils.NewDiscordSession(t)

	newCommand(t, d).Handle(s, slash("get"))

	assert.Empty(t, d.requests.calls)
	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	assert.True(t, strings.HasPrefix(*payload.Content, "Oups! Il semble qu'il n'y a pas de serveur "),
		*payload.Content)
}

func TestHandle_GetGuildLookupErrorPanics(t *testing.T) {
	d := newDeps()
	d.guilds.err = errors.New("db down")
	s, _ := testutils.NewDiscordSession(t)

	assert.PanicsWithError(t, "db down", func() { newCommand(t, d).Handle(s, slash("get")) })
	assert.Empty(t, d.requests.calls)
}

func TestHandle_GetMembersFailurePanics(t *testing.T) {
	d := newDeps()
	d.servers.servers = []entities.Server{djaul()}
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/members", Status: 403, Body: `{"message":"no"}`})

	assert.Panics(t, func() { newCommand(t, d).Handle(s, slash("get", stringOption("server", "djaul"))) })
	assert.Empty(t, d.requests.calls)
}

func TestHandle_SetRequestsUpdateForAuthor(t *testing.T) {
	d := newDeps()
	d.books.cities = []entities.City{brakmar()}
	d.books.orders = []entities.Order{heart()}
	d.servers.servers = []entities.Server{djaul()}
	s, fake := testutils.NewDiscordSession(t)

	newCommand(t, d).Handle(s, slash("set", stringOption("city", "brak"), stringOption("order", "coeur"),
		levelOption(42), stringOption("server", "djaul")))

	assert.Empty(t, fake.Requests())
	call := d.requests.single(t)
	assert.Equal(t, amqp.RabbitMQMessage_ALIGN_SET_REQUEST, call.message.Type)
	request := call.message.GetAlignSetRequest()
	assert.Equal(t, "author-id", request.GetUserId())
	assert.Equal(t, "brakmar", request.GetCityId())
	assert.Equal(t, "heart", request.GetOrderId())
	assert.Equal(t, "djaul", request.GetServerId())
	assert.Equal(t, int64(42), request.GetLevel())
	assert.Empty(t, call.properties)
}

func TestHandle_SetStopsAtFirstFailedCheck(t *testing.T) {
	d := newDeps()
	d.books.orders = []entities.Order{heart()}
	s, fake := testutils.NewDiscordSession(t)

	newCommand(t, d).Handle(s, slash("set", stringOption("city", "nowhere"), stringOption("order", "coeur"),
		levelOption(42), stringOption("server", "djaul")))

	assert.Empty(t, d.requests.calls)
	require.Len(t, fake.Requests(), 1)
	assert.Equal(t, "Aie! '*nowhere*' ne correspond à aucune cité que je connaisse ! :pensive:",
		*fake.LastEdit(t).Content)
	assert.Len(t, d.books.findCalls, 1, "order is never looked up")
}

func TestHandle_UserCommandRequestsTargetBeliefs(t *testing.T) {
	d := newDeps()
	d.guilds.server, d.guilds.found = djaul(), true
	s, fake := testutils.NewDiscordSession(t)
	member := &discordgo.Member{Nick: "Target"}
	user := &discordgo.User{ID: "u9", Username: "target"}

	newCommand(t, d).Handle(s, userCommand("u9", &discordgo.ApplicationCommandInteractionDataResolved{
		Members: map[string]*discordgo.Member{"u9": member},
		Users:   map[string]*discordgo.User{"u9": user},
	}))

	assert.Empty(t, fake.Requests())
	call := d.requests.single(t)
	assert.Equal(t, amqp.RabbitMQMessage_ALIGN_GET_USER_REQUEST, call.message.Type)
	assert.Equal(t, "u9", call.message.GetAlignGetUserRequest().GetUserId())
	assert.Equal(t, "djaul", call.message.GetAlignGetUserRequest().GetServerId())
	require.Len(t, call.properties, 1)
	got, ok := call.properties[0]["user"].(*discordgo.Member)
	require.True(t, ok)
	assert.Same(t, member, got)
	assert.Same(t, user, got.User, "resolved user is attached to the member")
}

func TestHandle_UserCommandWithoutResolvedTargetPanics(t *testing.T) {
	d := newDeps()
	d.guilds.server, d.guilds.found = djaul(), true
	s, _ := testutils.NewDiscordSession(t)

	assert.Panics(t, func() {
		newCommand(t, d).Handle(s, userCommand("u9", &discordgo.ApplicationCommandInteractionDataResolved{
			Users: map[string]*discordgo.User{"u9": {ID: "u9"}},
		}))
	})
	assert.Empty(t, d.requests.calls)
}

func TestHandle_UnknownSubCommandPanics(t *testing.T) {
	s, _ := testutils.NewDiscordSession(t)
	assert.PanicsWithValue(t, commands.ErrNoProvidedHandler, func() {
		newCommand(t, newDeps()).Handle(s, slash("delete"))
	})
}

func TestHandle_UnhandledInteractionIsIgnored(t *testing.T) {
	d := newDeps()
	s, fake := testutils.NewDiscordSession(t)
	userAutocomplete := userCommand("u9", nil)
	userAutocomplete.Type = discordgo.InteractionApplicationCommandAutocomplete

	cmd := newCommand(t, d)
	cmd.Handle(s, interaction(discordgo.InteractionModalSubmit, discordgo.ModalSubmitInteractionData{}))
	cmd.Handle(s, userAutocomplete)

	assert.Empty(t, d.requests.calls)
	assert.Empty(t, fake.Requests())
}

func TestHandle_RequestFailurePanics(t *testing.T) {
	events := map[string]*discordgo.InteractionCreate{
		"get": slash("get", stringOption("server", "djaul")),
		"set": slash("set", stringOption("city", "bonta"), stringOption("order", "spirit"),
			levelOption(1), stringOption("server", "djaul")),
		"user": userCommand("u9", &discordgo.ApplicationCommandInteractionDataResolved{
			Members: map[string]*discordgo.Member{"u9": {}},
			Users:   map[string]*discordgo.User{"u9": {ID: "u9"}},
		}),
		"page":  component("/books/align?city=_&order=_&server=djaul&page=1"),
		"city":  component("/books/align?order=_&server=djaul", "bonta"),
		"order": component("/books/align?city=_&server=djaul", "spirit"),
	}
	for name, event := range events {
		t.Run(name, func(t *testing.T) {
			d := newDeps()
			d.requests.err = errors.New("broker down")
			d.books.cities = []entities.City{bonta()}
			d.books.orders = []entities.Order{spirit()}
			d.servers.servers = []entities.Server{djaul()}
			d.guilds.server, d.guilds.found = djaul(), true
			s, fake := testutils.NewDiscordSession(t)
			respondMembers(fake)

			assert.PanicsWithError(t, "broker down", func() { newCommand(t, d).Handle(s, event) })
		})
	}
}

func TestOptionGetters_MissingContextValuesPanic(t *testing.T) {
	d := newDeps()
	cmd := newCommand(t, d)
	s, _ := testutils.NewDiscordSession(t)
	type entry struct {
		key   constants.ContextKey
		value any
	}
	all := []entry{
		{constants.ContextKeyCity, bonta()}, {constants.ContextKeyOrder, spirit()},
		{constants.ContextKeyLevel, int64(1)}, {constants.ContextKeyServer, djaul()},
	}

	assert.Panics(t, func() { cmd.getBook(context.Background(), s, slash("get"), nil) })
	assert.Panics(t, func() { cmd.userBook(context.Background(), s, userCommand("u9", nil), nil) })
	for index, missing := range all {
		ctx := context.Background()
		for other, present := range all {
			if other != index {
				ctx = context.WithValue(ctx, present.key, present.value)
			}
		}
		assert.Panics(t, func() { cmd.setBook(ctx, s, slash("set"), nil) }, "missing %v", missing.key)
	}
	assert.Empty(t, d.requests.calls)
}
