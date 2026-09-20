package item

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
	"github.com/kaellybot/kaelly-discord/services/equipments"
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

type fakeEquipments struct {
	equipments.Service
}

type fakeEmojis struct {
	emojis.Service
}

func (fakeEmojis) GetMiscStringEmoji(emojiID constants.EmojiMiscID) string { return string(emojiID) }
func (fakeEmojis) GetMiscEmoji(emojiID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: string(emojiID)}
}
func (fakeEmojis) GetEntityStringEmoji(string, constants.EmojiEntity) string { return "entity" }
func (fakeEmojis) GetEquipmentStringEmoji(amqp.EquipmentType) string         { return "equip" }
func (fakeEmojis) GetItemTypeStringEmoji(amqp.ItemType) string               { return "type" }
func (fakeEmojis) GetItemTypeEmoji(amqp.ItemType) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: "type"}
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
	viper.Set(constants.ItemID, "item-discord-id")
	t.Cleanup(func() { viper.Set(constants.ItemID, nil) })
	manager := &fakeRequestManager{}
	return New(fakeCharacteristics{}, fakeEquipments{}, fakeEmojis{}, manager), manager
}

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		Locale: discordgo.French, Data: data,
		Member: &discordgo.Member{User: &discordgo.User{ID: authorID}},
	}}
}

func queryOption(value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: contract.ItemQueryOptionName, Type: discordgo.ApplicationCommandOptionString, Value: value,
	}
}

func itemCommand(options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: contract.ItemCommandName, Options: options})
}

func itemAutocomplete(options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	event := itemCommand(options...)
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
	assert.Equal(t, constants.ItemRequestRoutingKey, call.routingKey)
	assert.Equal(t, authorID, call.message.UserID)
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
		"slash command": {itemCommand(), true},
		"autocomplete":  {itemAutocomplete(), true},
		"other command": {interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "set"}), false},
		"item select":      {component("/item?type=EQUIPMENT_TYPE"), true},
		"item effects":     {component("/item/123/effects?type=EQUIPMENT_TYPE"), true},
		"item recipe":      {component("/item/123/recipe?type=EQUIPMENT_TYPE"), true},
		"set component":    {component("/set/123"), false},
		"unrelated prefix": {component("/items?type=X"), false},
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
		Name:        "/item",
		CommandID:   "</item:item-discord-id>",
		Description: testutils.Translations(t, discordgo.French, "item.help.detailed")[0],
		TutorialURL: testutils.Translations(t, discordgo.French, "item.help.tutorial")[0],
	}, descriptions[0])
	assert.Equal(t, "item", cmd.GetName())
	assert.NotEqual(t, descriptions[0].TutorialURL, cmd.GetDescriptions(discordgo.EnglishGB)[0].TutorialURL)
}

// ---- slash command ----

func TestHandle_QueryRequestsItemByName(t *testing.T) {
	cmd, manager := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	cmd.Handle(s, itemCommand(queryOption("Coiffe du Bouftou")))

	call := singleRequest(t, manager)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_REQUEST, call.message.Type)
	req := call.message.GetEncyclopediaItemRequest()
	assert.Equal(t, "Coiffe du Bouftou", req.GetQuery())
	assert.False(t, req.GetIsID())
	assert.Equal(t, amqp.ItemType_ANY_ITEM_TYPE, req.GetType())
	assert.Empty(t, call.properties)
	assert.Empty(t, fake.Requests())
}

func TestHandle_BlankQueryAnswersConstraintError(t *testing.T) {
	for name, event := range map[string]*discordgo.InteractionCreate{
		"blank":   itemCommand(queryOption("   ")),
		"missing": itemCommand(),
	} {
		t.Run(name, func(t *testing.T) {
			cmd, manager := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			cmd.Handle(s, event)

			assert.Empty(t, manager.calls)
			req := fake.Last(t)
			assert.Equal(t, "PATCH", req.Method)
			assert.Equal(t, "/api/v9/webhooks/app/tok/messages/@original", req.Path)
			payload := fake.LastEdit(t)
			require.NotNil(t, payload.Content)
			assert.Equal(t, testutils.Translations(t, discordgo.French, "checks.query.constraints")[0],
				*payload.Content)
		})
	}
}

func TestHandle_RequestFailurePanics(t *testing.T) {
	cmd, manager := newCommand(t)
	manager.err = errors.New("broker down")
	s, _ := testutils.NewDiscordSession(t)

	assert.PanicsWithError(t, "broker down", func() { cmd.Handle(s, itemCommand(queryOption("x"))) })
	assert.PanicsWithError(t, "broker down", func() {
		cmd.Handle(s, component("/item/1/recipe?type=EQUIPMENT_TYPE"))
	})
}

func TestGetItem_MissingQueryInContextPanics(t *testing.T) {
	cmd, manager := newCommand(t)
	assert.Panics(t, func() { cmd.getItem(context.Background(), nil, itemCommand(), nil) })
	assert.Empty(t, manager.calls)
}

// ---- message components ----

func TestHandle_ComponentsRequestItemByID(t *testing.T) {
	cases := map[string]struct {
		event    *discordgo.InteractionCreate
		query    string
		itemType amqp.ItemType
		isRecipe bool
	}{
		"select from set": {component("/item?type=COSMETIC_TYPE", "42"), "42", amqp.ItemType_COSMETIC_TYPE, false},
		"effects button": {component("/item/77/effects?type=EQUIPMENT_TYPE"), "77",
			amqp.ItemType_EQUIPMENT_TYPE, false},
		"recipe button": {component("/item/88/recipe?type=EQUIPMENT_TYPE"), "88",
			amqp.ItemType_EQUIPMENT_TYPE, true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, manager := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			cmd.Handle(s, tc.event)

			call := singleRequest(t, manager)
			req := call.message.GetEncyclopediaItemRequest()
			assert.Equal(t, tc.query, req.GetQuery())
			assert.True(t, req.GetIsID())
			assert.Equal(t, tc.itemType, req.GetType())
			require.Len(t, call.properties, 1)
			assert.Equal(t, map[string]any{isRecipeProperty: tc.isRecipe}, call.properties[0])
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestHandle_InvalidComponentsPanic(t *testing.T) {
	cases := map[string]*discordgo.InteractionCreate{
		"select without value":   component("/item?type=EQUIPMENT_TYPE"),
		"select with two values": component("/item?type=EQUIPMENT_TYPE", "1", "2"),
		"unknown item type":      component("/item/1/effects?type=WEAPON"),
		"foreign custom ID":      component("/set/1"),
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, manager := newCommand(t)
			s, _ := testutils.NewDiscordSession(t)

			assert.PanicsWithError(t, commands.ErrInvalidInteraction.Error(), func() { cmd.Handle(s, event) })
			assert.Empty(t, manager.calls)
		})
	}
}

// ---- autocomplete ----

func TestHandle_AutocompleteRequestsItemList(t *testing.T) {
	cmd, manager := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	cmd.Handle(s, itemAutocomplete(focused(queryOption("bouf"))))

	call := singleRequest(t, manager)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, call.message.Type)
	assert.Equal(t, "bouf", call.message.GetEncyclopediaListRequest().GetQuery())
	assert.Equal(t, amqp.EncyclopediaListRequest_ITEM, call.message.GetEncyclopediaListRequest().GetType())
	assert.Empty(t, fake.Requests(), "the choices are sent when the answer arrives")
}

func TestHandle_AutocompleteIgnoresBlankUnfocusedOrUnknownOptions(t *testing.T) {
	cases := map[string]*discordgo.InteractionCreate{
		"blank query": itemAutocomplete(focused(queryOption("  "))),
		"not focused": itemAutocomplete(queryOption("bouf")),
		"unknown option": itemAutocomplete(focused(&discordgo.ApplicationCommandInteractionDataOption{
			Name: "other", Type: discordgo.ApplicationCommandOptionString, Value: "bouf"})),
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, manager := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			cmd.Handle(s, event)

			assert.Empty(t, manager.calls)
			assert.Empty(t, fake.Requests(), "no answer at all: Discord shows its own empty state")
		})
	}
}

func TestHandle_AutocompleteRequestFailureIsSwallowed(t *testing.T) {
	cmd, manager := newCommand(t)
	manager.err = errors.New("broker down")
	s, fake := testutils.NewDiscordSession(t)

	assert.NotPanics(t, func() { cmd.Handle(s, itemAutocomplete(focused(queryOption("bouf")))) })
	assert.Empty(t, fake.Requests())
}

type autocompletePayload struct {
	Type discordgo.InteractionResponseType `json:"type"`
	Data struct {
		Choices []*discordgo.ApplicationCommandOptionChoice `json:"choices"`
	} `json:"data"`
}

func TestAutocompleteItemList_AnswersChoicesFromList(t *testing.T) {
	cmd, _ := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	cmd.autocompleteItemList(context.Background(), s, itemAutocomplete(), &amqp.RabbitMQMessage{
		Type:   amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER,
		Status: amqp.RabbitMQMessage_SUCCESS,
		EncyclopediaListAnswer: &amqp.EncyclopediaListAnswer{Items: []*amqp.EncyclopediaListAnswer_Item{
			{Id: "1", Name: "Coiffe du Bouftou"}, {Id: "2", Name: "Cape du Bouftou"},
		}},
	}, nil)

	req := fake.Last(t)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/api/v9/interactions/interaction-id/tok/callback", req.Path)
	var payload autocompletePayload
	req.Decode(t, &payload)
	assert.Equal(t, discordgo.InteractionApplicationCommandAutocompleteResult, payload.Type)
	assert.Equal(t, []*discordgo.ApplicationCommandOptionChoice{
		{Name: "Coiffe du Bouftou", Value: "Coiffe du Bouftou"},
		{Name: "Cape du Bouftou", Value: "Cape du Bouftou"},
	}, payload.Data.Choices, "item names are both label and value, in answer order")
}

func TestAutocompleteItemList_InvalidAnswerSendsEmptyChoices(t *testing.T) {
	items := &amqp.EncyclopediaListAnswer{Items: []*amqp.EncyclopediaListAnswer_Item{{Name: "x"}}}
	cases := map[string]*amqp.RabbitMQMessage{
		"failed": {Type: amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER, Status: amqp.RabbitMQMessage_FAILED,
			EncyclopediaListAnswer: items},
		"wrong type": {Type: amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER, Status: amqp.RabbitMQMessage_SUCCESS,
			EncyclopediaListAnswer: items},
		"missing answer": {Type: amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER, Status: amqp.RabbitMQMessage_SUCCESS},
	}
	for name, message := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, _ := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)
			fake.Respond(testutils.DiscordResponse{PathContains: "/callback", Status: 400, Body: `{}`})

			assert.NotPanics(t, func() {
				cmd.autocompleteItemList(context.Background(), s, itemAutocomplete(), message, nil)
			})

			var payload autocompletePayload
			fake.Last(t).Decode(t, &payload)
			assert.Equal(t, discordgo.InteractionApplicationCommandAutocompleteResult, payload.Type)
			assert.Empty(t, payload.Data.Choices)
		})
	}
}

// ---- item answers ----

func itemAnswer(equipment *amqp.EncyclopediaItemAnswer_Equipment) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER,
		Status:   amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		EncyclopediaItemAnswer: &amqp.EncyclopediaItemAnswer{
			Type:      amqp.ItemType_EQUIPMENT_TYPE,
			Query:     "bouftou",
			Equipment: equipment,
			Source:    &amqp.Source{Name: "dofusdude"},
		},
	}
}

func equipment() *amqp.EncyclopediaItemAnswer_Equipment {
	return &amqp.EncyclopediaItemAnswer_Equipment{
		Id: "8", Name: "Bouftou Hat", Level: 10,
		Effects: []*amqp.EncyclopediaItemAnswer_Effect{{Id: "vita", Label: "+10 Vitality"}},
		Recipe: &amqp.EncyclopediaItemAnswer_Recipe{Ingredients: []*amqp.EncyclopediaItemAnswer_Recipe_Ingredient{
			{Id: "wool", Name: "Bouftou Wool", Quantity: 3},
		}},
	}
}

type componentRow struct {
	Components []struct {
		CustomID string `json:"custom_id"`
	} `json:"components"`
}

func buttonIDs(t *testing.T, payload testutils.WebhookEditPayload) []string {
	t.Helper()
	ids := make([]string, 0)
	for _, raw := range payload.Components {
		var row componentRow
		require.NoError(t, discordgo.Unmarshal(raw, &row))
		for _, c := range row.Components {
			ids = append(ids, c.CustomID)
		}
	}
	return ids
}

func TestGetItemReply_ShowsEffectsWithRecipeButton(t *testing.T) {
	cmd, _ := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	cmd.getItemReply(context.Background(), s, itemCommand(), itemAnswer(equipment()), nil)

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "Bouftou Hat", embed.Title)
	require.NotEmpty(t, embed.Fields)
	assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "item.effects.title")[0], embed.Fields[0].Name,
		"answer language (EN) drives the rendering, and effects are shown first")
	assert.Equal(t, []string{"/item/8/recipe?type=EQUIPMENT_TYPE"}, buttonIDs(t, payload))
}

func TestUpdateItemReply_RecipePropertySwitchesView(t *testing.T) {
	cmd, _ := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	cmd.updateItemReply(context.Background(), s, component("/item/8/recipe?type=EQUIPMENT_TYPE"),
		itemAnswer(equipment()), map[string]any{isRecipeProperty: true})

	payload := fake.LastEdit(t)
	embed := payload.Embeds[0]
	require.NotEmpty(t, embed.Fields)
	assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "item.recipe.title")[0], embed.Fields[0].Name)
	assert.Contains(t, embed.Fields[0].Value, "Bouftou Wool")
	assert.Equal(t, []string{"/item/8/effects?type=EQUIPMENT_TYPE"}, buttonIDs(t, payload))
}

func TestUpdateItemReply_NoEquipmentAnswersQueryMismatch(t *testing.T) {
	cmd, _ := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	cmd.updateItemReply(context.Background(), s, itemCommand(), itemAnswer(nil), nil)

	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	assert.Contains(t, *payload.Content, "bouftou")
	assert.Empty(t, payload.Embeds)
	assertQueryMismatch(t, discordgo.EnglishGB, "bouftou", *payload.Content)
}

func assertQueryMismatch(t *testing.T, locale discordgo.Locale, query, content string) {
	t.Helper()
	expected := make([]string, 0)
	for _, variant := range testutils.Translations(t, locale, "errors.query_mismatch") {
		expected = append(expected, strings.ReplaceAll(variant, "{{ .value }}", query))
	}
	assert.Contains(t, expected, content)
}

func TestUpdateItemReply_InvalidAnswerPanics(t *testing.T) {
	failed := itemAnswer(equipment())
	failed.Status = amqp.RabbitMQMessage_FAILED
	wrongType := itemAnswer(equipment())
	wrongType.Type = amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER
	missing := itemAnswer(nil)
	missing.EncyclopediaItemAnswer = nil

	for name, message := range map[string]*amqp.RabbitMQMessage{
		"failed": failed, "wrong type": wrongType, "missing answer": missing,
	} {
		t.Run(name, func(t *testing.T) {
			cmd, _ := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			assert.PanicsWithError(t, commands.ErrInvalidAnswerMessage.Error(), func() {
				cmd.updateItemReply(context.Background(), s, itemCommand(), message,
					map[string]any{isRecipeProperty: false})
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestUpdateItemReply_MissingOrWrongRecipePropertyPanics(t *testing.T) {
	for name, properties := range map[string]map[string]any{
		"missing":    {},
		"not a bool": {isRecipeProperty: "true"},
	} {
		t.Run(name, func(t *testing.T) {
			cmd, _ := newCommand(t)
			s, fake := testutils.NewDiscordSession(t)

			assert.PanicsWithError(t, commands.ErrRequestPropertyNotFound.Error(), func() {
				cmd.updateItemReply(context.Background(), s, itemCommand(), itemAnswer(equipment()), properties)
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestHandle_CallbacksRenderMatchingView(t *testing.T) {
	cmd, manager := newCommand(t)
	s, fake := testutils.NewDiscordSession(t)

	cmd.Handle(s, itemCommand(queryOption("bouftou")))
	cmd.Handle(s, component("/item/8/recipe?type=EQUIPMENT_TYPE"))
	require.Len(t, manager.calls, 2)

	manager.calls[0].callback(context.Background(), s, itemCommand(), itemAnswer(equipment()), nil)
	assert.Equal(t, []string{"/item/8/recipe?type=EQUIPMENT_TYPE"}, buttonIDs(t, fake.LastEdit(t)),
		"slash command answers default to the effects view")

	call := manager.calls[1]
	call.callback(context.Background(), s, component("x"), itemAnswer(equipment()), call.properties[0])
	assert.Equal(t, []string{"/item/8/effects?type=EQUIPMENT_TYPE"}, buttonIDs(t, fake.LastEdit(t)))
}
