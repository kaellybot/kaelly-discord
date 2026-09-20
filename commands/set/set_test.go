package set

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
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authorID = "author-1"

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

type fakeCharacteristics struct {
	characteristics.Service
}

func (fakeCharacteristics) GetCharacteristic(id string) (entities.Characteristic, bool) {
	return entities.Characteristic{ID: id}, false
}

type fakeEmojis struct {
	emojis.Service
}

func (fakeEmojis) GetEntityStringEmoji(string, constants.EmojiEntity) string { return "entity" }
func (fakeEmojis) GetSetBonusEmoji(int) *discordgo.ComponentEmoji            { return nil }
func (fakeEmojis) GetEquipmentEmoji(amqp.EquipmentType) *discordgo.ComponentEmoji {
	return nil
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

func newCommand(t *testing.T) (*Command, *fakeRequestManager) {
	t.Helper()
	viper.Set(constants.SetID, "set-discord-id")
	t.Cleanup(func() { viper.Set(constants.SetID, nil) })
	manager := &fakeRequestManager{}
	return New(fakeCharacteristics{}, fakeEmojis{}, manager), manager
}

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		Locale: discordgo.French, Data: data,
		User: &discordgo.User{ID: authorID}, // DM usage: no member
	}}
}

func queryOption(value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: contract.SetQueryOptionName, Type: discordgo.ApplicationCommandOptionString, Value: value,
	}
}

func setCommand(options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: contract.SetCommandName, Options: options})
}

func setAutocomplete(options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	event := setCommand(options...)
	event.Type = discordgo.InteractionApplicationCommandAutocomplete
	return event
}

func component(customID string, values ...string) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: customID, Values: values})
}

func focused(option *discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.ApplicationCommandInteractionDataOption {
	option.Focused = true
	return option
}

func singleRequest(t *testing.T, manager *fakeRequestManager) requestCall {
	t.Helper()
	require.Len(t, manager.calls, 1)
	call := manager.calls[0]
	assert.Equal(t, constants.SetRequestRoutingKey, call.routingKey)
	assert.Equal(t, authorID, call.message.UserID, "user ID is taken from DM interactions too")
	assert.Equal(t, amqp.Language_FR, call.message.Language)
	return call
}

// ---- Matches / descriptions ----

func TestMatches(t *testing.T) {
	cmd, _ := newCommand(t)
	cases := map[string]struct {
		event *discordgo.InteractionCreate
		want  bool
	}{
		"slash command": {setCommand(), true},
		"autocomplete":  {setAutocomplete(), true},
		"other command": {interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "item"}), false},
		"set button":     {component("/set/123"), true},
		"bonus select":   {component("/set/123/bonuses"), true},
		"item component": {component("/item?type=EQUIPMENT_TYPE"), false},
		"unknown suffix": {component("/set/123/other"), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, cmd.Matches(tc.event))
		})
	}
}

func TestGetDescriptions(t *testing.T) {
	cmd, _ := newCommand(t)

	descriptions := cmd.GetDescriptions(discordgo.French)

	require.Len(t, descriptions, 1)
	assert.Equal(t, commands.Description{
		Name:        "/set",
		CommandID:   "</set:set-discord-id>",
		Description: testutils.Translations(t, discordgo.French, "set.help.detailed")[0],
		TutorialURL: testutils.Translations(t, discordgo.French, "set.help.tutorial")[0],
	}, descriptions[0])
	assert.Equal(t, "set", cmd.GetName())
}

// ---- slash command ----

func TestHandle_QueryRequestsSetByName(t *testing.T) {
	cmd, manager := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.Handle(s, setCommand(queryOption("Panoplie du Bouftou"))))

	call := singleRequest(t, manager)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_REQUEST, call.message.Type)
	req := call.message.GetEncyclopediaItemRequest()
	assert.Equal(t, "Panoplie du Bouftou", req.GetQuery())
	assert.False(t, req.GetIsID())
	assert.Equal(t, amqp.ItemType_SET_TYPE, req.GetType())
	assert.Empty(t, call.properties)
	assert.Empty(t, fake.Requests())
}

func TestHandle_BlankQueryAnswersConstraintError(t *testing.T) {
	for name, event := range map[string]*discordgo.InteractionCreate{
		"blank":   setCommand(queryOption("\t ")),
		"missing": setCommand(),
	} {
		t.Run(name, func(t *testing.T) {
			cmd, manager := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			require.NoError(t, cmd.Handle(s, event))

			assert.Empty(t, manager.calls)
			assert.Equal(t, "PATCH", fake.Last(t).Method)
			payload := fake.LastEdit(t)
			require.NotNil(t, payload.Content)
			assert.Equal(t, testutils.Translations(t, discordgo.French, "checks.query.constraints")[0],
				*payload.Content)
		})
	}
}

func TestHandle_RequestFailureIsReported(t *testing.T) {
	cmd, manager := newCommand(t)
	errBroker := errors.New("broker down")
	manager.err = errBroker
	s, fake := testutils.NewDiscordSession(t)

	require.ErrorIs(t, cmd.Handle(s, setCommand(queryOption("x"))), errBroker)
	require.ErrorIs(t, cmd.Handle(s, component("/set/1")), errBroker)
	assert.Empty(t, fake.Requests(), "the router answers the caller, not the command")
}

func TestGetSet_MissingQueryInContextIsReported(t *testing.T) {
	cmd, manager := newCommand(t)
	require.Error(t, cmd.getSet(context.Background(), nil, setCommand(), nil))
	assert.Empty(t, manager.calls)
}

// ---- message components ----

func TestHandle_SetButtonRequestsSetByIDWithDefaultView(t *testing.T) {
	cmd, manager := newCommand(t)
	s, _ := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.Handle(s, component("/set/250")))

	call := singleRequest(t, manager)
	req := call.message.GetEncyclopediaItemRequest()
	assert.Equal(t, "250", req.GetQuery())
	assert.True(t, req.GetIsID())
	assert.Equal(t, amqp.ItemType_SET_TYPE, req.GetType())
	require.Len(t, call.properties, 1)
	assert.Empty(t, call.properties[0])
}

func TestHandle_BonusSelectRequestsSetWithItemNumber(t *testing.T) {
	cmd, manager := newCommand(t)
	s, _ := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.Handle(s, component("/set/250/bonuses", "3")))

	call := singleRequest(t, manager)
	req := call.message.GetEncyclopediaItemRequest()
	assert.Equal(t, "250", req.GetQuery())
	assert.True(t, req.GetIsID())
	require.Len(t, call.properties, 1)
	assert.Equal(t, map[string]any{itemNumberProperty: 3}, call.properties[0])
}

func TestHandle_InvalidComponentsAreReportedAsStale(t *testing.T) {
	cases := map[string]struct {
		event    *discordgo.InteractionCreate
		contains string
	}{
		"bonus without value":  {component("/set/1/bonuses"), "/set/1/bonuses"},
		"bonus with 2 values":  {component("/set/1/bonuses", "1", "2"), "/set/1/bonuses"},
		"bonus not a number":   {component("/set/1/bonuses", "two"), `strconv.Atoi: parsing "two": invalid syntax`},
		"foreign custom ID":    {component("/item/1/recipe?type=EQUIPMENT_TYPE"), "/item/1/recipe"},
		"set ID with a dash":   {component("/set/1-2"), "/set/1-2"},
		"unknown set subroute": {component("/set/1/other"), "/set/1/other"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, manager := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			err := cmd.Handle(s, tc.event)

			require.ErrorIs(t, err, commands.ErrInvalidInteraction)
			assert.ErrorIs(t, err, commands.ErrStaleInteraction,
				"a component we cannot read is a stale click, not a crash")
			assert.ErrorContains(t, err, tc.contains)
			assert.Empty(t, manager.calls)
			assert.Empty(t, fake.Requests(), "the router answers the caller, not the command")
		})
	}
}

// ---- autocomplete ----

func TestHandle_AutocompleteRequestsSetList(t *testing.T) {
	cmd, manager := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.Handle(s, setAutocomplete(focused(queryOption("bouf")))))

	call := singleRequest(t, manager)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, call.message.Type)
	assert.Equal(t, "bouf", call.message.GetEncyclopediaListRequest().GetQuery())
	assert.Equal(t, amqp.EncyclopediaListRequest_SET, call.message.GetEncyclopediaListRequest().GetType())
	assert.Empty(t, fake.Requests())
}

func TestHandle_AutocompleteIgnoresBlankUnfocusedOrUnknownOptions(t *testing.T) {
	cases := map[string]*discordgo.InteractionCreate{
		"blank query": setAutocomplete(focused(queryOption(" "))),
		"not focused": setAutocomplete(queryOption("bouf")),
		"unknown option": setAutocomplete(focused(&discordgo.ApplicationCommandInteractionDataOption{
			Name: "other", Type: discordgo.ApplicationCommandOptionString, Value: "bouf"})),
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, manager := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			require.NoError(t, cmd.Handle(s, event))

			assert.Empty(t, manager.calls)
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestHandle_AutocompleteRequestFailureIsSwallowed(t *testing.T) {
	cmd, manager := newCommand(t)
	manager.err = errors.New("broker down")
	s, _ := testutils.NewDiscordSession(t)

	assert.NoError(t, cmd.Handle(s, setAutocomplete(focused(queryOption("bouf")))),
		"an autocompletion cannot be answered with an error message")
}

type autocompletePayload struct {
	Type discordgo.InteractionResponseType `json:"type"`
	Data struct {
		Choices []*discordgo.ApplicationCommandOptionChoice `json:"choices"`
	} `json:"data"`
}

func listAnswer(status amqp.RabbitMQMessage_Status, messageType amqp.RabbitMQMessage_Type,
	names ...string) *amqp.RabbitMQMessage {
	items := make([]*amqp.EncyclopediaListAnswer_Item, 0)
	for _, name := range names {
		items = append(items, &amqp.EncyclopediaListAnswer_Item{Id: "id-" + name, Name: name})
	}
	return &amqp.RabbitMQMessage{Type: messageType, Status: status,
		EncyclopediaListAnswer: &amqp.EncyclopediaListAnswer{Items: items}}
}

func TestAutocompleteSetList_AnswersChoicesFromList(t *testing.T) {
	cmd, _ := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.autocompleteSetList(context.Background(), s, setAutocomplete(),
		listAnswer(amqp.RabbitMQMessage_SUCCESS, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER,
			"Panoplie du Bouftou", "Panoplie du Boune"), nil))

	req := fake.Last(t)
	assert.Equal(t, "/api/v9/interactions/interaction-id/tok/callback", req.Path)
	var payload autocompletePayload
	req.Decode(t, &payload)
	assert.Equal(t, discordgo.InteractionApplicationCommandAutocompleteResult, payload.Type)
	assert.Equal(t, []*discordgo.ApplicationCommandOptionChoice{
		{Name: "Panoplie du Bouftou", Value: "Panoplie du Bouftou"},
		{Name: "Panoplie du Boune", Value: "Panoplie du Boune"},
	}, payload.Data.Choices)
}

func TestAutocompleteSetList_InvalidAnswerSendsEmptyChoices(t *testing.T) {
	missing := listAnswer(amqp.RabbitMQMessage_SUCCESS, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER)
	missing.EncyclopediaListAnswer = nil
	cases := map[string]*amqp.RabbitMQMessage{
		"failed": listAnswer(amqp.RabbitMQMessage_FAILED, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER, "x"),
		"wrong type": listAnswer(amqp.RabbitMQMessage_SUCCESS,
			amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER, "x"),
		"missing answer": missing,
	}
	for name, message := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, _ := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)
			fake.Respond(testutils.DiscordResponse{PathContains: "/callback", Status: 400, Body: `{}`})

			assert.NoError(t, cmd.autocompleteSetList(context.Background(), s, setAutocomplete(),
				message, nil))

			var payload autocompletePayload
			fake.Last(t).Decode(t, &payload)
			assert.Empty(t, payload.Data.Choices)
		})
	}
}

// ---- set answers ----

func setAnswer(set *amqp.EncyclopediaItemAnswer_Set) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER,
		Status:   amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		EncyclopediaItemAnswer: &amqp.EncyclopediaItemAnswer{
			Type:   amqp.ItemType_SET_TYPE,
			Query:  "bouftou",
			Set:    set,
			Source: &amqp.Source{Name: "dofusdude"},
		},
	}
}

func bouftouSet() *amqp.EncyclopediaItemAnswer_Set {
	return &amqp.EncyclopediaItemAnswer_Set{
		Id: "250", Name: "Bouftou Set", Level: 10,
		Equipments: []*amqp.EncyclopediaItemAnswer_Set_Equipment{
			{Id: "1", Name: "Bouftou Hat"}, {Id: "2", Name: "Bouftou Cloak"}, {Id: "3", Name: "Bouftou Ring"},
		},
		Bonuses: []*amqp.EncyclopediaItemAnswer_Set_Bonus{
			{ItemNumber: 2, Effects: []*amqp.EncyclopediaItemAnswer_Effect{{Id: "a", Label: "+10 Agility"}}},
			{ItemNumber: 3, Effects: []*amqp.EncyclopediaItemAnswer_Effect{{Id: "s", Label: "+20 Strength"}}},
		},
	}
}

type selectRow struct {
	Components []struct {
		CustomID string `json:"custom_id"`
		Options  []struct {
			Value   string `json:"value"`
			Default bool   `json:"default"`
		} `json:"options"`
	} `json:"components"`
}

// selectedBonus returns the default option value of the bonus select menu.
func selectedBonus(t *testing.T, payload testutils.WebhookEditPayload) string {
	t.Helper()
	for _, raw := range payload.Components {
		var row selectRow
		require.NoError(t, discordgo.Unmarshal(raw, &row))
		for _, c := range row.Components {
			if c.CustomID != "/set/250/bonuses" {
				continue
			}
			for _, option := range c.Options {
				if option.Default {
					return option.Value
				}
			}
		}
	}
	return ""
}

func bonusFieldText(payload testutils.WebhookEditPayload) string {
	var text strings.Builder
	for _, field := range payload.Embeds[0].Fields {
		text.WriteString(field.Name + "|" + field.Value + "\n")
	}
	return text.String()
}

func TestGetSetReply_DefaultViewShowsFullSetBonus(t *testing.T) {
	cmd, _ := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.getSetReply(context.Background(), s, setCommand(),
		setAnswer(bouftouSet()), nil))

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	assert.Equal(t, "Bouftou Set", payload.Embeds[0].Title)
	assert.Equal(t, "3", selectedBonus(t, payload), "all 3 items equipped by default")
	assert.Contains(t, bonusFieldText(payload), "Strength")
	assert.NotContains(t, bonusFieldText(payload), "Agility")
}

func TestUpdateSetReply_ShowsRequestedBonus(t *testing.T) {
	cmd, _ := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.updateSetReply(context.Background(), s, component("/set/250/bonuses", "2"),
		setAnswer(bouftouSet()), map[string]any{itemNumberProperty: 2}))

	payload := fake.LastEdit(t)
	assert.Equal(t, "2", selectedBonus(t, payload))
	assert.Contains(t, bonusFieldText(payload), "Agility")
	assert.NotContains(t, bonusFieldText(payload), "Strength")
}

func TestUpdateSetReply_MissingOrWrongItemNumberIsReported(t *testing.T) {
	for name, properties := range map[string]map[string]any{
		"missing":      {},
		"not an int":   {itemNumberProperty: "2"},
		"int64 is not": {itemNumberProperty: int64(2)},
	} {
		t.Run(name, func(t *testing.T) {
			cmd, _ := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			err := cmd.updateSetReply(context.Background(), s, component("/set/250/bonuses", "2"),
				setAnswer(bouftouSet()), properties)

			require.ErrorIs(t, err, commands.ErrRequestPropertyNotFound)
			assert.NotErrorIs(t, err, commands.ErrStaleInteraction,
				"wiring gone wrong is a bug, not a stale click")
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestGetSetReply_NoSetAnswersQueryMismatch(t *testing.T) {
	cmd, _ := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.getSetReply(context.Background(), s, setCommand(), setAnswer(nil), nil))

	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	expected := make([]string, 0)
	for _, variant := range testutils.Translations(t, discordgo.EnglishGB, "errors.query_mismatch") {
		expected = append(expected, strings.ReplaceAll(variant, "{{ .value }}", "bouftou"))
	}
	assert.Contains(t, expected, *payload.Content, "answer language drives the message")
	assert.Empty(t, payload.Embeds)
}

func TestSetReplies_InvalidAnswerIsReported(t *testing.T) {
	failed := setAnswer(bouftouSet())
	failed.Status = amqp.RabbitMQMessage_FAILED
	wrongType := setAnswer(bouftouSet())
	wrongType.Type = amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER
	missing := setAnswer(nil)
	missing.EncyclopediaItemAnswer = nil
	notASet := setAnswer(bouftouSet())
	notASet.EncyclopediaItemAnswer.Type = amqp.ItemType_EQUIPMENT_TYPE

	for name, message := range map[string]*amqp.RabbitMQMessage{
		"failed": failed, "wrong type": wrongType, "missing answer": missing, "item is not a set": notASet,
	} {
		t.Run(name, func(t *testing.T) {
			cmd, _ := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)
			properties := map[string]any{itemNumberProperty: 2}

			for _, reply := range []requests.RequestCallback{cmd.getSetReply, cmd.updateSetReply} {
				err := reply(context.Background(), s, setCommand(), message, properties)

				require.ErrorIs(t, err, commands.ErrInvalidAnswerMessage)
				assert.ErrorIs(t, err, commands.ErrUnavailable, "a backend that failed is worth a retry")
			}
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestHandle_CallbacksMatchComponentKind(t *testing.T) {
	cmd, manager := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.Handle(s, setCommand(queryOption("bouftou"))))
	require.NoError(t, cmd.Handle(s, component("/set/250")))
	require.NoError(t, cmd.Handle(s, component("/set/250/bonuses", "2")))
	require.Len(t, manager.calls, 3)

	for i, want := range []string{"3", "3", "2"} {
		call := manager.calls[i]
		var properties map[string]any
		if len(call.properties) > 0 {
			properties = call.properties[0]
		}
		require.NoError(t, call.callback(context.Background(), s, setCommand(),
			setAnswer(bouftouSet()), properties))
		assert.Equal(t, want, selectedBonus(t, fake.LastEdit(t)), "call #%d", i)
	}
}
