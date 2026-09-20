package almanax

import (
	"context"
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
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

type fakeEmojis struct {
	emojis.Service
}

func (fakeEmojis) GetMiscEmoji(emojiID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: string(emojiID)}
}

func (fakeEmojis) GetMiscStringEmoji(emojiID constants.EmojiMiscID) string {
	return ":" + string(emojiID) + ":"
}

func (fakeEmojis) GetItemTypeStringEmoji(itemType amqp.ItemType) string {
	return ":" + itemType.String() + ":"
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

// single asserts exactly one request was emitted to the encyclopedia routing key.
func (f *fakeRequestManager) single(t *testing.T) requestCall {
	t.Helper()
	require.Len(t, f.calls, 1)
	assert.Equal(t, "requests.encyclopedias", f.calls[0].routingKey)
	assert.Equal(t, "author-id", f.calls[0].message.UserID)
	assert.Equal(t, amqp.Language_FR, f.calls[0].message.Language)
	return f.calls[0]
}

func newCommand(t *testing.T, manager requests.RequestManager) *Command {
	t.Helper()
	viper.Set(constants.AlmanaxID, "almanax-discord-id")
	t.Cleanup(func() { viper.Set(constants.AlmanaxID, nil) })
	return New(fakeEmojis{}, manager)
}

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		Locale: discordgo.French, Data: data,
		Member: &discordgo.Member{User: &discordgo.User{ID: "author-id"}},
	}}
}

func slash(subCommand string, options ...*discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand, discordgo.ApplicationCommandInteractionData{
		Name: "almanax",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{{
			Name: subCommand, Type: discordgo.ApplicationCommandOptionSubCommand, Options: options,
		}},
	})
}

func autocompleteEvent(options ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	event := slash("effects", options...)
	event.Type = discordgo.InteractionApplicationCommandAutocomplete
	return event
}

func stringOption(name, value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: name, Type: discordgo.ApplicationCommandOptionString, Value: value,
	}
}

func durationOption(value int64) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: "duration", Type: discordgo.ApplicationCommandOptionInteger, Value: float64(value),
	}
}

func component(customID string, values ...string) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: customID, Values: values})
}

func TestGetName_IsAlmanax(t *testing.T) {
	assert.Equal(t, "almanax", newCommand(t, &fakeRequestManager{}).GetName())
}

func TestGetDescriptions_OneEntryPerSubCommand(t *testing.T) {
	descriptions := newCommand(t, &fakeRequestManager{}).GetDescriptions(discordgo.French)

	require.Len(t, descriptions, 3)
	for index, sub := range []string{"day", "effects", "resources"} {
		assert.Equal(t, "/almanax "+sub, descriptions[index].Name)
		assert.Equal(t, "</almanax "+sub+":almanax-discord-id>", descriptions[index].CommandID)
		assert.Equal(t, testutils.Translations(t, discordgo.French, "almanax.help.detailed."+sub)[0],
			descriptions[index].Description)
		assert.Equal(t, testutils.Translations(t, discordgo.French, "almanax.help.tutorial."+sub)[0],
			descriptions[index].TutorialURL)
	}
}

func TestMatches(t *testing.T) {
	cmd := newCommand(t, &fakeRequestManager{})
	cases := []struct {
		name     string
		event    *discordgo.InteractionCreate
		expected bool
	}{
		{"slash command", slash("day"), true},
		{"autocomplete", autocompleteEvent(), true},
		{"other slash command", interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "map"}), false},
		{"day button", component("/almanax/day/1700000000"), true},
		{"day choice", component("/almanax/day?type=choice"), true},
		{"effect page", component("/almanax/effect?date=1700000000&page=1"), true},
		{"resource duration", component("/almanax/resource?characters=2"), true},
		{"resource character", component("/almanax/resource?duration=7"), true},
		{"foreign button", component("/map/1?type=normal"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, cmd.Matches(tc.event))
		})
	}
}

func TestHandle_SubCommandsRunTheirChecksThenRequest(t *testing.T) {
	cases := map[string]struct {
		event  *discordgo.InteractionCreate
		verify func(t *testing.T, message *amqp.RabbitMQMessage)
	}{
		"day": {slash("day", stringOption("date", "18/09/2024")), func(t *testing.T, message *amqp.RabbitMQMessage) {
			assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_REQUEST, message.Type)
			assert.Equal(t, "2024-09-18T00:00:00Z",
				message.GetEncyclopediaAlmanaxRequest().GetDate().AsTime().Format("2006-01-02T15:04:05Z07:00"))
		}},
		"resources": {slash("resources", durationOption(12)), func(t *testing.T, message *amqp.RabbitMQMessage) {
			assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_REQUEST, message.Type)
			assert.Equal(t, int64(12), message.GetEncyclopediaAlmanaxResourceRequest().GetDuration())
		}},
		"effects": {slash("effects", stringOption("effect", "Pêche")), func(t *testing.T, message *amqp.RabbitMQMessage) {
			assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_REQUEST, message.Type)
			request := message.GetEncyclopediaAlmanaxEffectRequest()
			assert.Equal(t, amqp.EncyclopediaAlmanaxEffectRequest_QUERY, request.GetType())
			assert.Equal(t, "Pêche", request.GetQuery())
			assert.Equal(t, int64(0), request.GetOffset())
			assert.Equal(t, int64(constants.MaxAlmanaxEffectPerEmbed), request.GetSize())
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			manager := &fakeRequestManager{}
			s, fake := testutils.NewDiscordSession(t)

			require.NoError(t, newCommand(t, manager).Handle(s, tc.event))

			assert.Empty(t, fake.Requests(), "no error answer on valid input")
			call := manager.single(t)
			assert.Empty(t, call.properties)
			tc.verify(t, call.message)
		})
	}
}

func TestHandle_FailedCheckStopsBeforeRequest(t *testing.T) {
	manager := &fakeRequestManager{}
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, newCommand(t, manager).Handle(s, slash("resources", durationOption(99))))

	assert.Empty(t, manager.calls)
	assert.Len(t, fake.Requests(), 1)
}

func TestHandle_UnknownSubCommandIsReported(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)

	err := newCommand(t, &fakeRequestManager{}).Handle(s, slash("unknown"))

	require.ErrorIs(t, err, commands.ErrNoProvidedHandler)
	assert.ErrorIs(t, err, commands.ErrStaleInteraction,
		"a sub-command nobody handles is a stale interaction, not a crash")
	assert.Empty(t, fake.Requests(), "the router answers the caller, not the command")
}

func TestHandle_UnknownComponentIsReported(t *testing.T) {
	s, fake := testutils.NewDiscordSession(t)

	err := newCommand(t, &fakeRequestManager{}).Handle(s, component("/almanax/unknown"))

	require.ErrorIs(t, err, commands.ErrNoProvidedHandler)
	assert.ErrorIs(t, err, commands.ErrStaleInteraction,
		"a component nobody handles is a stale click, not a crash")
	assert.Empty(t, fake.Requests(), "the router answers the caller, not the command")
}

func TestHandle_UnhandledInteractionTypeIsIgnored(t *testing.T) {
	manager := &fakeRequestManager{}
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, newCommand(t, manager).Handle(s, interaction(discordgo.InteractionModalSubmit,
		discordgo.ModalSubmitInteractionData{CustomID: "/almanax/day/1"})))

	assert.Empty(t, manager.calls)
	assert.Empty(t, fake.Requests())
}

func TestHandle_RequestFailureIsReportedForEverySubCommand(t *testing.T) {
	events := map[string]*discordgo.InteractionCreate{
		"day":       slash("day"),
		"resources": slash("resources"),
		"effects":   slash("effects", stringOption("effect", "xp")),
	}
	for name, event := range events {
		t.Run(name, func(t *testing.T) {
			s, fake := testutils.NewDiscordSession(t)
			errBroker := errors.New("broker down")
			cmd := newCommand(t, &fakeRequestManager{err: errBroker})

			require.ErrorIs(t, cmd.Handle(s, event), errBroker)
			assert.Empty(t, fake.Requests(), "the router answers the caller, not the command")
		})
	}
}

func TestRequestHandlers_MissingContextValueIsReported(t *testing.T) {
	manager := &fakeRequestManager{}
	cmd := newCommand(t, manager)
	s, _ := testutils.NewDiscordSession(t)
	ctx := context.Background()

	require.Error(t, cmd.getAlmanax(ctx, s, slash("day"), nil))
	require.Error(t, cmd.getResources(ctx, s, slash("resources"), nil))
	require.Error(t, cmd.getAlmanaxesByEffect(ctx, s, slash("effects"), nil))
	assert.Empty(t, manager.calls)
}
