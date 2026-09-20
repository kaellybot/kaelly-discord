package align

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- checks ---------------------------------------------------------------

type checkResult struct {
	ctx    context.Context
	called bool
	fake   *testutils.FakeDiscord
}

func (r checkResult) content(t *testing.T) string {
	t.Helper()
	require.Len(t, r.fake.Requests(), 1)
	payload := r.fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	return *payload.Content
}

func runCheck(t *testing.T, d *deps, check func(*Command) middlewares.MiddlewareCommand,
	event *discordgo.InteractionCreate) checkResult {
	t.Helper()
	cmd := newCommand(t, d)
	s, fake := testutils.NewDiscordSession(t)
	result := checkResult{fake: fake}
	check(cmd)(context.Background(), s, event, func(ctx context.Context) {
		result.called = true
		result.ctx = ctx
	})
	return result
}

func mandatoryCity(cmd *Command) middlewares.MiddlewareCommand  { return cmd.checkMandatoryCity }
func optionalCity(cmd *Command) middlewares.MiddlewareCommand   { return cmd.checkOptionalCity }
func mandatoryOrder(cmd *Command) middlewares.MiddlewareCommand { return cmd.checkMandatoryOrder }
func optionalOrder(cmd *Command) middlewares.MiddlewareCommand  { return cmd.checkOptionalOrder }
func level(cmd *Command) middlewares.MiddlewareCommand          { return cmd.checkLevel }

func TestCheckMandatoryCity_SingleMatchIsForwarded(t *testing.T) {
	d := newDeps()
	d.books.cities = []entities.City{bonta()}

	result := runCheck(t, d, mandatoryCity, slash("set", stringOption("city", "bon")))

	require.True(t, result.called)
	assert.Equal(t, bonta(), result.ctx.Value(constants.ContextKeyCity))
	assert.Equal(t, []findCall{{"bon", discordgo.French, constants.MaxChoices}}, d.books.findCalls)
	assert.Empty(t, result.fake.Requests())
}

func TestCheckMandatoryCity_NoOrManyMatchesAnswersError(t *testing.T) {
	cases := map[string]struct {
		cities   []entities.City
		expected string
	}{
		"none": {nil, "Aie! '*b*' ne correspond à aucune cité que je connaisse ! :pensive:"},
		"many": {[]entities.City{bonta(), brakmar()},
			"Outch! '*b*' correspond à plusieurs cités (Bonta, Brâkmar) ! :slight_smile:"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := newDeps()
			d.books.cities = tc.cities

			result := runCheck(t, d, mandatoryCity, slash("set", stringOption("city", "b")))

			assert.False(t, result.called)
			assert.Equal(t, tc.expected, result.content(t))
		})
	}
}

func TestCheckMandatoryOrder_SingleMatchIsForwarded(t *testing.T) {
	d := newDeps()
	d.books.orders = []entities.Order{heart()}

	result := runCheck(t, d, mandatoryOrder, slash("set", stringOption("order", "coe")))

	require.True(t, result.called)
	assert.Equal(t, heart(), result.ctx.Value(constants.ContextKeyOrder))
	assert.Empty(t, result.fake.Requests())
}

func TestCheckMandatoryOrder_NoOrManyMatchesAnswersError(t *testing.T) {
	cases := map[string]struct {
		orders   []entities.Order
		expected string
	}{
		"none": {nil, "Aie! '*x*' ne correspond à aucun ordre que je connaisse ! :pensive:"},
		"many": {[]entities.Order{spirit(), heart()},
			"Outch! '*x*' correspond à plusieurs ordres (Esprit, Coeur) ! :slight_smile:"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := newDeps()
			d.books.orders = tc.orders

			result := runCheck(t, d, mandatoryOrder, slash("set", stringOption("order", "x")))

			assert.False(t, result.called)
			assert.Equal(t, tc.expected, result.content(t))
		})
	}
}

// Discord enforces the mandatory options, but an absent one must still be
// answered: the interaction is deferred and would otherwise hang.
func TestCheckMandatory_AbsentOptionAnswersEmptyFieldError(t *testing.T) {
	for name, check := range map[string]func(*Command) middlewares.MiddlewareCommand{
		"city": mandatoryCity, "order": mandatoryOrder, "level": level,
	} {
		t.Run(name, func(t *testing.T) {
			result := runCheck(t, newDeps(), check, slash("set"))

			assert.False(t, result.called)
			assert.Equal(t, "Le champs ne doit pas être vide ! :grin:", result.content(t))
		})
	}
}

func TestCheckOptionalCityAndOrder_OnlyAnExactSingleMatchIsForwarded(t *testing.T) {
	cases := map[string]struct {
		cities    []entities.City
		orders    []entities.Order
		event     *discordgo.InteractionCreate
		wantCity  any
		wantOrder any
	}{
		"single matches": {[]entities.City{bonta()}, []entities.Order{spirit()},
			slash("get", stringOption("city", "b"), stringOption("order", "e")), bonta(), spirit()},
		"absent": {[]entities.City{bonta()}, []entities.Order{spirit()}, slash("get"), nil, nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := newDeps()
			d.books.cities, d.books.orders = tc.cities, tc.orders

			city := runCheck(t, d, optionalCity, tc.event)
			order := runCheck(t, d, optionalOrder, tc.event)

			require.True(t, city.called)
			require.True(t, order.called)
			assert.Equal(t, tc.wantCity, city.ctx.Value(constants.ContextKeyCity))
			assert.Equal(t, tc.wantOrder, order.ctx.Value(constants.ContextKeyOrder))
			assert.Empty(t, city.fake.Requests())
			assert.Empty(t, order.fake.Requests())
		})
	}
}

// An optional filter that matches nothing, or several entries, is a typo: it is
// answered like the mandatory checks do, instead of silently listing the whole
// book.
func TestCheckOptionalCityAndOrder_AmbiguousOrUnknownAnswersError(t *testing.T) {
	cases := map[string]struct {
		prepare  func(*deps)
		check    func(*Command) middlewares.MiddlewareCommand
		expected string
	}{
		"unknown city": {func(*deps) {}, optionalCity,
			"Aie! '*b*' ne correspond à aucune cité que je connaisse ! :pensive:"},
		"ambiguous city": {func(d *deps) { d.books.cities = []entities.City{bonta(), brakmar()} }, optionalCity,
			"Outch! '*b*' correspond à plusieurs cités (Bonta, Brâkmar) ! :slight_smile:"},
		"unknown order": {func(*deps) {}, optionalOrder,
			"Aie! '*e*' ne correspond à aucun ordre que je connaisse ! :pensive:"},
		"ambiguous order": {func(d *deps) { d.books.orders = []entities.Order{spirit(), heart()} }, optionalOrder,
			"Outch! '*e*' correspond à plusieurs ordres (Esprit, Coeur) ! :slight_smile:"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := newDeps()
			tc.prepare(d)

			result := runCheck(t, d, tc.check,
				slash("get", stringOption("city", "b"), stringOption("order", "e")))

			assert.False(t, result.called, "the book is not listed unfiltered")
			assert.Equal(t, tc.expected, result.content(t))
		})
	}
}

func TestCheckLevel(t *testing.T) {
	for _, value := range []int64{0, 100} {
		result := runCheck(t, newDeps(), level, slash("set", levelOption(value)))
		require.True(t, result.called)
		assert.Equal(t, value, result.ctx.Value(constants.ContextKeyLevel))
	}
	for _, value := range []int64{-1, 101} {
		result := runCheck(t, newDeps(), level, slash("set", levelOption(value)))
		assert.False(t, result.called)
		assert.Equal(t, "Le niveau doit être compris entre 0 et 100 ! :grin:", result.content(t))
	}
}

func TestChecks_DiscordFailureIsSwallowed(t *testing.T) {
	cases := map[string]struct {
		check func(*Command) middlewares.MiddlewareCommand
		event *discordgo.InteractionCreate
	}{
		"city":  {mandatoryCity, slash("set", stringOption("city", "x"))},
		"order": {mandatoryOrder, slash("set", stringOption("order", "x"))},
		"level": {level, slash("set", levelOption(500))},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := newCommand(t, newDeps())
			s, fake := testutils.NewDiscordSession(t)
			fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 500, Body: `{}`})

			assert.NotPanics(t, func() {
				tc.check(cmd)(context.Background(), s, tc.event,
					func(context.Context) { t.Fatal("next must not be called") })
			})
			assert.Len(t, fake.Requests(), 1)
		})
	}
}

// --- autocomplete ---------------------------------------------------------

type autocompleteResponse struct {
	Type discordgo.InteractionResponseType `json:"type"`
	Data struct {
		Choices []discordgo.ApplicationCommandOptionChoice `json:"choices"`
	} `json:"data"`
}

func autocompleteEvent(options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	event := slash("get", options...)
	event.Type = discordgo.InteractionApplicationCommandAutocomplete
	return event
}

func focused(option *discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.ApplicationCommandInteractionDataOption {
	option.Focused = true
	return option
}

func autocomplete(t *testing.T, d *deps, event *discordgo.InteractionCreate) autocompleteResponse {
	t.Helper()
	s, fake := testutils.NewDiscordSession(t)
	newCommand(t, d).Handle(s, event)

	req := fake.Last(t)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, discordgo.EndpointInteractionResponse("interaction-id", "tok"), "https://discord.com"+req.Path)
	var response autocompleteResponse
	req.Decode(t, &response)
	assert.Equal(t, discordgo.InteractionApplicationCommandAutocompleteResult, response.Type)
	return response
}

func choiceNames(t *testing.T, response autocompleteResponse) []string {
	t.Helper()
	names := make([]string, 0, len(response.Data.Choices))
	for _, choice := range response.Data.Choices {
		assert.Equal(t, choice.Name, choice.Value, "label is used as both name and value")
		names = append(names, choice.Name)
	}
	return names
}

func TestAutocomplete_FocusedOptionDrivesTheLookup(t *testing.T) {
	d := newDeps()
	d.books.cities = []entities.City{bonta(), brakmar()}
	d.books.orders = []entities.Order{spirit(), heart()}
	d.servers.servers = []entities.Server{djaul(), hel()}

	cities := autocomplete(t, d, autocompleteEvent(stringOption("order", "x"), focused(stringOption("city", "b"))))
	orders := autocomplete(t, d, autocompleteEvent(focused(stringOption("order", "e"))))
	servers := autocomplete(t, d, autocompleteEvent(focused(stringOption("server", "h"))))

	assert.Equal(t, []string{"Bonta", "Brâkmar"}, choiceNames(t, cities))
	assert.Equal(t, []string{"Esprit", "Coeur"}, choiceNames(t, orders))
	assert.Equal(t, []string{"Djaul", "Hel Munster"}, choiceNames(t, servers))
	assert.Equal(t, []findCall{
		{"b", discordgo.French, constants.MaxChoices},
		{"e", discordgo.French, constants.MaxChoices},
	}, d.books.findCalls)
	assert.Equal(t, []findCall{{"h", discordgo.French, constants.MaxChoices}}, d.servers.findCalls)
}

func TestAutocomplete_LabelFallsBackToDefaultLocaleThenID(t *testing.T) {
	d := newDeps()
	d.servers.servers = []entities.Server{
		{ID: "en-only", Labels: []entities.ServerLabel{{Locale: amqp.Language_EN, Label: "English"}}},
		{ID: "no-label"},
	}
	event := autocompleteEvent(focused(stringOption("server", "x")))
	event.Locale = discordgo.German

	response := autocomplete(t, d, event)

	assert.Equal(t, []string{"English", "no-label"}, choiceNames(t, response))
}

func TestAutocomplete_NoFocusedOrUnknownOptionRespondsWithoutChoices(t *testing.T) {
	d := newDeps()
	d.books.cities = []entities.City{bonta()}

	for name, event := range map[string]*discordgo.InteractionCreate{
		"nothing focused": autocompleteEvent(stringOption("city", "b")),
		"unknown option":  autocompleteEvent(focused(stringOption("level", "1"))),
	} {
		t.Run(name, func(t *testing.T) {
			response := autocomplete(t, d, event)
			assert.Empty(t, response.Data.Choices)
		})
	}
	assert.Empty(t, d.books.findCalls)
}

func TestAutocomplete_DiscordFailureIsSwallowed(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/callback", Status: 500, Body: `{}`})

	assert.NotPanics(t, func() {
		newCommand(t, newDeps()).Handle(s, autocompleteEvent(focused(stringOption("city", "b"))))
	})
	assert.Len(t, fake.Requests(), 1)
}

// --- book components ------------------------------------------------------

func TestComponents_RequestTheBookForTheSelectedFilters(t *testing.T) {
	cases := map[string]struct {
		event               *discordgo.InteractionCreate
		city, order, server string
		expectedOffset      int64
	}{
		"page button": {component("/books/align?city=bonta&order=_&server=djaul&page=2"), "bonta", "", "djaul", 30},
		"city selected": {component("/books/align?order=heart&server=djaul", "brakmar"),
			"brakmar", "heart", "djaul", 0},
		"all cities selected": {component("/books/align?order=heart&server=djaul", "_"), "", "heart", "djaul", 0},
		"order selected":      {component("/books/align?city=bonta&server=hel", "spirit"), "bonta", "spirit", "hel", 0},
		"all orders selected": {component("/books/align?city=_&server=hel", "_"), "", "", "hel", 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := newDeps()
			s, fake := testutils.NewDiscordSession(t)
			respondMembers(fake)

			newCommand(t, d).Handle(s, tc.event)

			call := d.requests.single(t)
			assert.Equal(t, amqp.RabbitMQMessage_ALIGN_GET_BOOK_REQUEST, call.message.Type)
			request := call.message.GetAlignGetBookRequest()
			assert.Equal(t, tc.city, request.GetCityId())
			assert.Equal(t, tc.order, request.GetOrderId())
			assert.Equal(t, tc.server, request.GetServerId())
			assert.Equal(t, tc.expectedOffset, request.GetOffset())
			assert.ElementsMatch(t, []string{"u1", "u2"}, request.GetUserIds())
			require.Len(t, call.properties, 1)
			assert.Equal(t, map[string]any{"u1": "Ali", "u2": "bob"}, call.properties[0])
		})
	}
}

func TestComponents_InvalidSelectionOrCustomIDPanics(t *testing.T) {
	d := newDeps()
	cmd := newCommand(t, d)
	s, fake := testutils.NewDiscordSession(t)
	respondMembers(fake)

	cases := map[string]struct {
		handler commands.DiscordHandler
		event   *discordgo.InteractionCreate
	}{
		"page: bad custom id":  {cmd.updateBookPage, component("/books/align?page=x")},
		"city: no value":       {cmd.updateCityBook, component("/books/align?order=_&server=djaul")},
		"city: two values":     {cmd.updateCityBook, component("/books/align?order=_&server=djaul", "a", "b")},
		"city: bad custom id":  {cmd.updateCityBook, component("/books/align?nope", "bonta")},
		"order: no value":      {cmd.updateOrderBook, component("/books/align?city=_&server=djaul")},
		"order: bad custom id": {cmd.updateOrderBook, component("/books/align?nope", "spirit")},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.PanicsWithValue(t, commands.ErrInvalidInteraction, func() { tc.handler(s, tc.event) })
		})
	}
	assert.Empty(t, d.requests.calls)
	assert.Empty(t, fake.Requests(), "members are not fetched for invalid interactions")
}

func TestComponents_MembersFailurePanics(t *testing.T) {
	for name, event := range map[string]*discordgo.InteractionCreate{
		"page":  component("/books/align?city=_&order=_&server=djaul&page=1"),
		"city":  component("/books/align?order=_&server=djaul", "bonta"),
		"order": component("/books/align?city=_&server=djaul", "spirit"),
	} {
		t.Run(name, func(t *testing.T) {
			d := newDeps()
			s, fake := testutils.NewDiscordSession(t)
			fake.Respond(testutils.DiscordResponse{PathContains: "/members", Status: 403, Body: `{}`})

			assert.Panics(t, func() { newCommand(t, d).Handle(s, event) })
			assert.Empty(t, d.requests.calls)
		})
	}
}

func TestComponents_UnknownCustomIDPanics(t *testing.T) {
	s, _ := testutils.NewDiscordSession(t)
	assert.PanicsWithValue(t, commands.ErrNoProvidedHandler, func() {
		newCommand(t, newDeps()).Handle(s, component("/books/align?unknown=1"))
	})
}

// --- replies --------------------------------------------------------------

func bookAnswer() *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type: amqp.RabbitMQMessage_ALIGN_GET_BOOK_ANSWER, Status: amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		AlignGetBookAnswer: &amqp.AlignGetBookAnswer{
			CityId: "", OrderId: "spirit", ServerId: "djaul", Page: 0, Pages: 1, Total: 3,
			Believers: []*amqp.AlignGetBookAnswer_Believer{
				{CityId: "bonta", OrderId: "spirit", UserId: "u1", Level: 50},
				{CityId: "brakmar", OrderId: "spirit", UserId: "gone", Level: 99},
				{CityId: "bonta", OrderId: "spirit", UserId: "u2", Level: 80},
			},
		},
	}
}

func bookDeps() *deps {
	d := newDeps()
	d.books.cities = []entities.City{bonta(), brakmar()}
	d.books.orders = []entities.Order{spirit(), heart()}
	d.servers.servers = []entities.Server{djaul()}
	return d
}

func TestGetBookReply_ListsOnlyKnownMembersSortedByLevel(t *testing.T) {
	cmd := newCommand(t, bookDeps())
	s, fake := testutils.NewDiscordSession(t)

	cmd.getBookReply(context.Background(), s, slash("get"), bookAnswer(),
		map[string]any{"u1": "Ali", "u2": "bob"})

	req := fake.Last(t)
	assert.Equal(t, "PATCH", req.Method)
	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "Aligned list", embed.Title, "answer language drives the reply locale")
	assert.Equal(t, "*3 results, page 1 on 1*\n\n"+
		":bonta: :lightspirit: lvl. 80: **bob**\n"+
		":bonta: :lightspirit: lvl. 50: **Ali**\n", embed.Description,
		"members missing from the guild are skipped")
	assert.Equal(t, bonta().Color, embed.Color, "the city with the most levels colors the embed")
	assert.Equal(t, "Djaul EN", embed.Footer.Text)

	type selectRow struct {
		Components []struct {
			CustomID string `json:"custom_id"`
			Options  []struct {
				Value   string `json:"value"`
				Default bool   `json:"default"`
			} `json:"options"`
		} `json:"components"`
	}
	rows := make([]selectRow, 0, len(payload.Components))
	for _, raw := range payload.Components {
		var row selectRow
		require.NoError(t, json.Unmarshal(raw, &row))
		rows = append(rows, row)
	}
	require.Len(t, rows, 2, "single page: no pagination row")
	assert.Equal(t, "/books/align?order=spirit&server=djaul", rows[0].Components[0].CustomID)
	assert.Equal(t, "/books/align?city=_&server=djaul", rows[1].Components[0].CustomID)
	assert.True(t, rows[0].Components[0].Options[0].Default, "all cities is selected")
	assert.Equal(t, "spirit", rows[1].Components[0].Options[1].Value)
	assert.True(t, rows[1].Components[0].Options[1].Default)
}

func TestGetBookReply_InvalidAnswerPanics(t *testing.T) {
	failed := bookAnswer()
	failed.Status = amqp.RabbitMQMessage_FAILED
	wrongType := bookAnswer()
	wrongType.Type = amqp.RabbitMQMessage_ALIGN_GET_USER_ANSWER
	missing := bookAnswer()
	missing.AlignGetBookAnswer = nil

	for name, message := range map[string]*amqp.RabbitMQMessage{
		"failed": failed, "wrong type": wrongType, "missing payload": missing,
	} {
		t.Run(name, func(t *testing.T) {
			s, fake := testutils.NewDiscordSession(t)
			assert.PanicsWithValue(t, commands.ErrInvalidAnswerMessage, func() {
				newCommand(t, bookDeps()).getBookReply(context.Background(), s, slash("get"), message, nil)
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestSetBookReply(t *testing.T) {
	t.Run("success answers in the answer language", func(t *testing.T) {
		s, fake := testutils.NewDiscordSession(t)
		newCommand(t, newDeps()).setBookReply(context.Background(), s, slash("set"),
			&amqp.RabbitMQMessage{Status: amqp.RabbitMQMessage_SUCCESS, Language: amqp.Language_FR}, nil)

		payload := fake.LastEdit(t)
		require.NotNil(t, payload.Content)
		assert.Equal(t, "L'annuaire a été mise à jour !", *payload.Content)
	})

	t.Run("failure panics", func(t *testing.T) {
		s, fake := testutils.NewDiscordSession(t)
		assert.PanicsWithValue(t, commands.ErrInvalidAnswerMessage, func() {
			newCommand(t, newDeps()).setBookReply(context.Background(), s, slash("set"),
				&amqp.RabbitMQMessage{Status: amqp.RabbitMQMessage_FAILED}, nil)
		})
		assert.Empty(t, fake.Requests())
	})
}

func userAnswer() *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type: amqp.RabbitMQMessage_ALIGN_GET_USER_ANSWER, Status: amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		AlignGetUserAnswer: &amqp.AlignGetUserAnswer{
			ServerId: "djaul",
			Beliefs: []*amqp.AlignGetUserAnswer_AlignExperience{
				{CityId: "brakmar", OrderId: "heart", Level: 20},
				{CityId: "bonta", OrderId: "spirit", Level: 40},
			},
		},
	}
}

func TestUserRespond_EditsWithMemberBeliefs(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)
	member := &discordgo.Member{Nick: "Target", User: &discordgo.User{ID: "u9", Username: "target"}}

	newCommand(t, bookDeps()).userRespond(context.Background(), s, userCommand("u9", nil), userAnswer(),
		map[string]any{"user": member})

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "Target alignments", embed.Title)
	assert.Equal(t, ":bonta: :lightspirit: lvl. 40\n:brakmar: :darkheart: lvl. 20\n", embed.Description)
	assert.Equal(t, bonta().Color, embed.Color)
	assert.Equal(t, "Djaul EN", embed.Footer.Text)
}

func TestUserRespond_FailureOrMissingMemberPanics(t *testing.T) {
	failed := userAnswer()
	failed.Status = amqp.RabbitMQMessage_FAILED
	member := &discordgo.Member{User: &discordgo.User{ID: "u9"}}

	cases := map[string]struct {
		message    *amqp.RabbitMQMessage
		properties map[string]any
	}{
		"failed status":  {failed, map[string]any{"user": member}},
		"missing member": {userAnswer(), map[string]any{}},
		"not a *Member":  {userAnswer(), map[string]any{"user": *member}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, fake := testutils.NewDiscordSession(t)
			assert.Panics(t, func() {
				newCommand(t, bookDeps()).userRespond(context.Background(), s, userCommand("u9", nil),
					tc.message, tc.properties)
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestReplies_DiscordFailureIsSwallowed(t *testing.T) {
	cmd := newCommand(t, bookDeps())
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 500, Body: `{}`})
	ctx := context.Background()
	member := &discordgo.Member{User: &discordgo.User{ID: "u9", Username: "target"}}

	assert.NotPanics(t, func() {
		cmd.getBookReply(ctx, s, slash("get"), bookAnswer(), map[string]any{})
		cmd.setBookReply(ctx, s, slash("set"), &amqp.RabbitMQMessage{Status: amqp.RabbitMQMessage_SUCCESS}, nil)
		cmd.userRespond(ctx, s, userCommand("u9", nil), userAnswer(), map[string]any{"user": member})
	})
	assert.Len(t, fake.Requests(), 3)
}
