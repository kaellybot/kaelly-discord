package almanax

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func refDay() time.Time { return time.Date(2024, 9, 18, 0, 0, 0, 0, time.UTC) }

func source() *amqp.Source {
	return &amqp.Source{Name: "Krosmoz", Url: "https://krosmoz", Icon: "https://krosmoz/icon"}
}

func almanaxAnswer() *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type: amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_ANSWER, Status: amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		EncyclopediaAlmanaxAnswer: &amqp.EncyclopediaAlmanaxAnswer{
			Source: source(),
			Almanax: &amqp.Almanax{
				Bonus:  "More XP",
				Reward: 1234,
				Date:   timestamppb.New(refDay()),
				Tribute: &amqp.Almanax_Tribute{Quantity: 3, Item: &amqp.Almanax_Tribute_Item{
					Name: "Wheat", Icon: "https://wheat.png", Type: amqp.ItemType_CONSUMABLE_TYPE,
				}},
			},
		},
	}
}

func effectAnswer() *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type: amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_ANSWER, Status: amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		EncyclopediaAlmanaxEffectAnswer: &amqp.EncyclopediaAlmanaxEffectAnswer{
			Query: "xp", EffectName: "More XP", Page: 0, Pages: 1, Total: 2, Source: source(),
			Almanaxes: []*amqp.Almanax{
				{Bonus: "More XP", Date: timestamppb.New(refDay())},
				{Bonus: "More XP again", Date: timestamppb.New(refDay().AddDate(0, 1, 0))},
			},
		},
	}
}

func resourceAnswer() *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type: amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_ANSWER, Status: amqp.RabbitMQMessage_SUCCESS,
		Language: amqp.Language_EN,
		EncyclopediaAlmanaxResourceAnswer: &amqp.EncyclopediaAlmanaxResourceAnswer{
			Duration: 7, Source: source(),
			Tributes: []*amqp.EncyclopediaAlmanaxResourceAnswer_Tribute{
				{ItemName: "Wheat", ItemType: amqp.ItemType_CONSUMABLE_TYPE, Quantity: 4},
			},
		},
	}
}

func unix(date time.Time) string {
	return strconv.FormatInt(date.Unix(), 10)
}

// invalidVariants returns the answer with a failed status, a wrong type and a missing payload.
func invalidVariants(build func() *amqp.RabbitMQMessage, clearPayload func(*amqp.RabbitMQMessage),
) map[string]*amqp.RabbitMQMessage {
	failed := build()
	failed.Status = amqp.RabbitMQMessage_FAILED
	wrongType := build()
	wrongType.Type = amqp.RabbitMQMessage_ABOUT_REQUEST
	missing := build()
	clearPayload(missing)
	return map[string]*amqp.RabbitMQMessage{"failed": failed, "wrong type": wrongType, "missing payload": missing}
}

func TestUpdateAlmanax_DayButtonRequestsThatDay(t *testing.T) {
	manager := &fakeRequestManager{}
	s, _ := testutils.NewDiscordSession(t)

	newCommand(t, manager).Handle(s, component("/almanax/day/"+unix(refDay())))

	call := manager.single(t)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_REQUEST, call.message.Type)
	assert.Equal(t, refDay(), call.message.GetEncyclopediaAlmanaxRequest().GetDate().AsTime())
}

func TestUpdateAlmanaxByDate_SelectedValueIsUnixSeconds(t *testing.T) {
	manager := &fakeRequestManager{}
	s, _ := testutils.NewDiscordSession(t)

	newCommand(t, manager).Handle(s, component("/almanax/day?type=choice", unix(refDay())))

	call := manager.single(t)
	assert.Equal(t, refDay(), call.message.GetEncyclopediaAlmanaxRequest().GetDate().AsTime())
}

func TestUpdateAlmanaxByDate_InvalidSelectionPanics(t *testing.T) {
	for name, values := range map[string][]string{
		"no value":     nil,
		"two values":   {"1", "2"},
		"not a number": {"today"},
	} {
		t.Run(name, func(t *testing.T) {
			manager := &fakeRequestManager{}
			s, _ := testutils.NewDiscordSession(t)

			assert.Panics(t, func() {
				newCommand(t, manager).Handle(s, component("/almanax/day?type=choice", values...))
			})
			assert.Empty(t, manager.calls)
		})
	}
}

func TestUpdateAlmanaxesByEffect_PageButtonRequestsByDate(t *testing.T) {
	manager := &fakeRequestManager{}
	s, _ := testutils.NewDiscordSession(t)

	newCommand(t, manager).Handle(s, component("/almanax/effect?date="+unix(refDay())+"&page=2"))

	call := manager.single(t)
	request := call.message.GetEncyclopediaAlmanaxEffectRequest()
	require.NotNil(t, request)
	assert.Equal(t, amqp.EncyclopediaAlmanaxEffectRequest_DATE, request.GetType())
	assert.Equal(t, refDay(), request.GetDate().AsTime())
	assert.Empty(t, request.GetQuery())
	assert.Equal(t, int64(10), request.GetOffset(), "page 2 with 5 effects per page")
	assert.Equal(t, int64(5), request.GetSize())
}

func TestUpdateResourceCharacter_KeepsDurationFromCustomIDAndCharactersFromSelection(t *testing.T) {
	manager := &fakeRequestManager{}
	s, _ := testutils.NewDiscordSession(t)

	newCommand(t, manager).Handle(s, component("/almanax/resource?duration=14", "50"))

	call := manager.single(t)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_REQUEST, call.message.Type)
	assert.Equal(t, int64(14), call.message.GetEncyclopediaAlmanaxResourceRequest().GetDuration())
	require.Len(t, call.properties, 1)
	assert.Equal(t, map[string]any{"characterNumber": int64(50)}, call.properties[0])
}

func TestUpdateResourceDuration_KeepsCharactersFromCustomIDAndDurationFromSelection(t *testing.T) {
	manager := &fakeRequestManager{}
	s, _ := testutils.NewDiscordSession(t)

	newCommand(t, manager).Handle(s, component("/almanax/resource?characters=4", "30"))

	call := manager.single(t)
	assert.Equal(t, int64(30), call.message.GetEncyclopediaAlmanaxResourceRequest().GetDuration())
	require.Len(t, call.properties, 1)
	assert.Equal(t, map[string]any{"characterNumber": int64(4)}, call.properties[0])
}

// The contract clamps the character number carried by the duration select to
// the *duration* bounds [1, 30], so a user tracking 50 characters falls back to
// 30 as soon as they change the period. Locked in as current behavior.
func TestUpdateResourceDuration_CharacterNumberIsClampedToDurationBounds(t *testing.T) {
	manager := &fakeRequestManager{}
	s, _ := testutils.NewDiscordSession(t)

	newCommand(t, manager).Handle(s, component("/almanax/resource?characters=50", "7"))

	call := manager.single(t)
	assert.Equal(t, map[string]any{"characterNumber": int64(30)}, call.properties[0])
}

func TestUpdateResources_InvalidSelectionPanics(t *testing.T) {
	for _, customID := range []string{"/almanax/resource?duration=14", "/almanax/resource?characters=4"} {
		t.Run(customID, func(t *testing.T) {
			manager := &fakeRequestManager{}
			s, _ := testutils.NewDiscordSession(t)

			assert.Panics(t, func() { newCommand(t, manager).Handle(s, component(customID, "x")) })
			assert.Empty(t, manager.calls)
		})
	}
}

func TestComponentHandlers_MalformedCustomIDPanics(t *testing.T) {
	manager := &fakeRequestManager{}
	cmd := newCommand(t, manager)
	s, _ := testutils.NewDiscordSession(t)
	event := component("/nope", "1")

	for name, handler := range map[string]commands.DiscordHandler{
		"day":                cmd.updateAlmanax,
		"effect":             cmd.updateAlmanaxesByEffect,
		"resource character": cmd.updateResourceCharacter,
		"resource duration":  cmd.updateResourceDuration,
	} {
		t.Run(name, func(t *testing.T) {
			assert.PanicsWithValue(t, commands.ErrInvalidInteraction, func() { handler(s, event) })
		})
	}
	assert.Empty(t, manager.calls)
}

func TestComponentHandlers_RequestFailurePanics(t *testing.T) {
	events := map[string]*discordgo.InteractionCreate{
		"day":                component("/almanax/day/" + unix(refDay())),
		"day choice":         component("/almanax/day?type=choice", unix(refDay())),
		"effect":             component("/almanax/effect?date=" + unix(refDay()) + "&page=0"),
		"resource character": component("/almanax/resource?duration=7", "2"),
		"resource duration":  component("/almanax/resource?characters=2", "7"),
	}
	for name, event := range events {
		t.Run(name, func(t *testing.T) {
			s, _ := testutils.NewDiscordSession(t)
			cmd := newCommand(t, &fakeRequestManager{err: errors.New("broker down")})
			assert.PanicsWithError(t, "broker down", func() { cmd.Handle(s, event) })
		})
	}
}

func TestAlmanaxRespond_EditsWithDayEmbedAndNavigation(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)

	cmd.almanaxRespond(context.Background(), s, slash("day"), almanaxAnswer(), nil)

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "https://www.krosmoz.com/en/almanax/2024-09-18", embed.URL, "answer language drives locale")
	assert.Equal(t, "https://wheat.png", embed.Image.URL)
	require.Len(t, embed.Fields, 3)
	assert.Equal(t, "More XP", embed.Fields[0].Value)
	assert.Equal(t, ":CONSUMABLE_TYPE: **3**x Wheat", embed.Fields[1].Value)
	assert.Contains(t, embed.Fields[2].Value, ":kama:")

	assert.Equal(t, []string{
		"/almanax/day/" + unix(refDay().AddDate(0, 0, -1)),
		"/almanax/day/" + unix(refDay().AddDate(0, 0, 1)),
		"/almanax/effect?date=" + unix(refDay()) + "&page=0",
	}, componentValues(t, payload, "custom_id"))
}

func TestAlmanaxRespond_MissingDayAnswersMissingMessage(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)
	answer := almanaxAnswer()
	answer.EncyclopediaAlmanaxAnswer.Almanax = nil

	cmd.almanaxRespond(context.Background(), s, slash("day"), answer, nil)

	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "almanax.day.missing")[0], *payload.Content)
}

func TestReplies_InvalidAnswersPanicWithoutAnswering(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	type reply func(context.Context, *discordgo.Session, *discordgo.InteractionCreate,
		*amqp.RabbitMQMessage, map[string]any)
	cases := map[string]struct {
		reply    reply
		variants map[string]*amqp.RabbitMQMessage
	}{
		"day": {cmd.almanaxRespond, invalidVariants(almanaxAnswer,
			func(m *amqp.RabbitMQMessage) { m.EncyclopediaAlmanaxAnswer = nil })},
		"effect": {cmd.effectRespond, invalidVariants(effectAnswer,
			func(m *amqp.RabbitMQMessage) { m.EncyclopediaAlmanaxEffectAnswer = nil })},
		"resource": {cmd.getResourcesReply, invalidVariants(resourceAnswer,
			func(m *amqp.RabbitMQMessage) { m.EncyclopediaAlmanaxResourceAnswer = nil })},
	}
	for name, tc := range cases {
		for variant, message := range tc.variants {
			t.Run(name+"/"+variant, func(t *testing.T) {
				s, fake := testutils.NewDiscordSession(t)
				assert.PanicsWithValue(t, commands.ErrInvalidAnswerMessage, func() {
					tc.reply(context.Background(), s, slash("day"), message, nil)
				})
				assert.Empty(t, fake.Requests())
			})
		}
	}
}

func TestEffectRespond_ListsDaysWithDayChoiceMenu(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)

	cmd.effectRespond(context.Background(), s, slash("effects"), effectAnswer(), nil)

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	assert.Equal(t, "xp", payload.Embeds[0].Title)
	require.Len(t, payload.Embeds[0].Fields, 2)
	assert.Equal(t, ":calendar: <t:"+unix(refDay())+":D>", payload.Embeds[0].Fields[0].Name)
	assert.Equal(t, []string{"/almanax/day?type=choice"}, componentValues(t, payload, "custom_id"),
		"single page: no pagination buttons")
	assert.Equal(t, []string{unix(refDay()), unix(refDay().AddDate(0, 1, 0))}, componentValues(t, payload, "value"))
}

func TestEffectRespond_UnknownEffectAndNoUpcomingDay(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})

	t.Run("no upcoming day", func(t *testing.T) {
		s, fake := testutils.NewDiscordSession(t)
		answer := effectAnswer()
		answer.EncyclopediaAlmanaxEffectAnswer.Almanaxes = nil

		cmd.effectRespond(context.Background(), s, slash("effects"), answer, nil)

		payload := fake.LastEdit(t)
		require.NotNil(t, payload.Content)
		assert.Equal(t, testutils.Translations(t, discordgo.EnglishGB, "almanax.effect.missing")[0], *payload.Content)
	})

	t.Run("unknown effect", func(t *testing.T) {
		s, fake := testutils.NewDiscordSession(t)
		answer := effectAnswer()
		answer.EncyclopediaAlmanaxEffectAnswer.EffectName = ""

		cmd.effectRespond(context.Background(), s, slash("effects"), answer, nil)

		payload := fake.LastEdit(t)
		require.NotNil(t, payload.Content)
		assert.Contains(t, *payload.Content, "xp", "the query is echoed back")
		assert.Empty(t, payload.Embeds)
	})
}

func TestGetResourcesReply_DefaultsToOneCharacter(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)

	cmd.getResourcesReply(context.Background(), s, slash("resources"), resourceAnswer(), nil)

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	assert.Equal(t, "- :CONSUMABLE_TYPE: **4**x Wheat\n", payload.Embeds[0].Description)
	assert.Equal(t, []string{"/almanax/resource?characters=1", "/almanax/resource?duration=7"},
		componentValues(t, payload, "custom_id"))
}

func TestUpdateResourcesReply_MultipliesQuantitiesByCharacterNumber(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)

	cmd.updateResourcesReply(context.Background(), s, component("/almanax/resource?duration=7", "3"),
		resourceAnswer(), map[string]any{"characterNumber": int64(3)})

	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	assert.Equal(t, "- :CONSUMABLE_TYPE: **12**x Wheat\n", payload.Embeds[0].Description)
	assert.Contains(t, componentValues(t, payload, "custom_id"), "/almanax/resource?characters=3")
}

func TestUpdateResourcesReply_MissingOrMistypedPropertyPanics(t *testing.T) {
	for name, properties := range map[string]map[string]any{
		"missing":   {},
		"not int64": {"characterNumber": 3},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := newCommand(t, &fakeRequestManager{})
			s, fake := testutils.NewDiscordSession(t)

			assert.PanicsWithValue(t, commands.ErrRequestPropertyNotFound, func() {
				cmd.updateResourcesReply(context.Background(), s, slash("resources"), resourceAnswer(), properties)
			})
			assert.Empty(t, fake.Requests())
		})
	}
}

func TestReplies_DiscordFailureIsSwallowed(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 500, Body: `{}`})
	ctx := context.Background()

	assert.NotPanics(t, func() {
		cmd.almanaxRespond(ctx, s, slash("day"), almanaxAnswer(), nil)
		cmd.effectRespond(ctx, s, slash("effects"), effectAnswer(), nil)
		cmd.getResourcesReply(ctx, s, slash("resources"), resourceAnswer(), nil)
	})
	assert.Len(t, fake.Requests(), 3)
}

// componentValues collects every string found under key in the sent components
// (e.g. "custom_id" or "value"), whatever the nesting.
func componentValues(t *testing.T, payload testutils.WebhookEditPayload, key string) []string {
	t.Helper()
	var values []string
	var walk func(node any)
	walk = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			for k, v := range typed {
				if text, ok := v.(string); ok && k == key {
					values = append(values, text)
				}
				walk(v)
			}
		case []any:
			for _, v := range typed {
				walk(v)
			}
		}
	}
	for _, raw := range payload.Components {
		var node any
		require.NoError(t, json.Unmarshal(raw, &node))
		walk(node)
	}
	return values
}
