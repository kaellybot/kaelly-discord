package almanax

import (
	"context"
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type autocompleteResponse struct {
	Type discordgo.InteractionResponseType `json:"type"`
	Data struct {
		Choices []discordgo.ApplicationCommandOptionChoice `json:"choices"`
	} `json:"data"`
}

func focused(option *discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.ApplicationCommandInteractionDataOption {
	option.Focused = true
	return option
}

func listAnswer(names ...string) *amqp.RabbitMQMessage {
	items := make([]*amqp.EncyclopediaListAnswer_Item, 0, len(names))
	for _, name := range names {
		items = append(items, &amqp.EncyclopediaListAnswer_Item{Id: "id-" + name, Name: name})
	}
	return &amqp.RabbitMQMessage{
		Type: amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER, Status: amqp.RabbitMQMessage_SUCCESS,
		EncyclopediaListAnswer: &amqp.EncyclopediaListAnswer{Items: items},
	}
}

func TestAutocomplete_FocusedEffectRequestsEffectList(t *testing.T) {
	manager := &fakeRequestManager{}
	s, fake := testutils.NewDiscordSession(t)

	newCommand(t, manager).Handle(s, autocompleteEvent(focused(stringOption("effect", "pêc"))))

	assert.Empty(t, fake.Requests(), "the answer is sent once the list comes back")
	call := manager.single(t)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, call.message.Type)
	assert.Equal(t, "pêc", call.message.GetEncyclopediaListRequest().GetQuery())
	assert.Equal(t, amqp.EncyclopediaListRequest_ALMANAX_EFFECT, call.message.GetEncyclopediaListRequest().GetType())
}

func TestAutocomplete_NonFocusedOrUnknownOptionsDoNothing(t *testing.T) {
	events := map[string]*discordgo.InteractionCreate{
		"not focused":    autocompleteEvent(stringOption("effect", "pêc")),
		"unknown option": autocompleteEvent(focused(stringOption("date", "18"))),
	}
	for name, event := range events {
		t.Run(name, func(t *testing.T) {
			manager := &fakeRequestManager{}
			s, fake := testutils.NewDiscordSession(t)

			newCommand(t, manager).Handle(s, event)

			assert.Empty(t, manager.calls)
			assert.Empty(t, fake.Requests())
		})
	}
}

// A blank query neither requests nor answers: Discord is left without an
// autocomplete response for that keystroke. Locked in as current behavior.
func TestAutocomplete_BlankQueryIsNotRequested(t *testing.T) {
	manager := &fakeRequestManager{}
	s, fake := testutils.NewDiscordSession(t)

	newCommand(t, manager).Handle(s, autocompleteEvent(focused(stringOption("effect", "  "))))

	assert.Empty(t, manager.calls)
	assert.Empty(t, fake.Requests())
}

func TestAutocomplete_RequestFailureIsSwallowed(t *testing.T) {
	manager := &fakeRequestManager{err: errors.New("broker down")}
	s, fake := testutils.NewDiscordSession(t)

	assert.NotPanics(t, func() {
		newCommand(t, manager).Handle(s, autocompleteEvent(focused(stringOption("effect", "xp"))))
	})
	assert.Len(t, manager.calls, 1)
	assert.Empty(t, fake.Requests())
}

func TestAutocompleteAlmanaxEffectList_RespondsWithEffectNames(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)
	event := autocompleteEvent(focused(stringOption("effect", "xp")))

	cmd.autocompleteAlmanaxEffectList(context.Background(), s, event, listAnswer("XP bonus", "XP malus"), nil)

	req := fake.Last(t)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, discordgo.EndpointInteractionResponse("interaction-id", "tok"), "https://discord.com"+req.Path)
	var response autocompleteResponse
	req.Decode(t, &response)
	assert.Equal(t, discordgo.InteractionApplicationCommandAutocompleteResult, response.Type)
	require.Len(t, response.Data.Choices, 2)
	assert.Equal(t, "XP bonus", response.Data.Choices[0].Name)
	assert.Equal(t, "XP bonus", response.Data.Choices[0].Value, "the name is used as value, not the id")
	assert.Equal(t, "XP malus", response.Data.Choices[1].Name)
}

func TestAutocompleteAlmanaxEffectList_InvalidAnswerRespondsWithoutChoices(t *testing.T) {
	failed := listAnswer("XP")
	failed.Status = amqp.RabbitMQMessage_FAILED
	wrongType := listAnswer("XP")
	wrongType.Type = amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_ANSWER
	missing := listAnswer()
	missing.EncyclopediaListAnswer = nil

	for name, message := range map[string]*amqp.RabbitMQMessage{
		"failed": failed, "wrong type": wrongType, "missing payload": missing,
	} {
		t.Run(name, func(t *testing.T) {
			cmd := newCommand(t, &fakeRequestManager{})
			s, fake := testutils.NewDiscordSession(t)

			cmd.autocompleteAlmanaxEffectList(context.Background(), s,
				autocompleteEvent(focused(stringOption("effect", "xp"))), message, nil)

			var response autocompleteResponse
			fake.Last(t).Decode(t, &response)
			assert.Equal(t, discordgo.InteractionApplicationCommandAutocompleteResult, response.Type)
			assert.Empty(t, response.Data.Choices)
		})
	}
}

func TestAutocompleteAlmanaxEffectList_DiscordFailureIsSwallowed(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/callback", Status: 500, Body: `{}`})

	assert.NotPanics(t, func() {
		cmd.autocompleteAlmanaxEffectList(context.Background(), s,
			autocompleteEvent(focused(stringOption("effect", "xp"))), listAnswer("XP"), nil)
	})
	assert.Len(t, fake.Requests(), 1)
}
