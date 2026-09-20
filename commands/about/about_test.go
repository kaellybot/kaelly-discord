package about

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
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
	requested []constants.EmojiMiscID
}

func (f *fakeEmojis) GetMiscStringEmoji(emojiID constants.EmojiMiscID) string {
	f.requested = append(f.requested, emojiID)
	return "<:" + string(emojiID) + ":1>"
}

func newCommand(t *testing.T, broker amqp.MessageBroker, emojiService emojis.Service) *Command {
	t.Helper()
	viper.Set(constants.AboutID, "about-discord-id")
	t.Cleanup(func() { viper.Set(constants.AboutID, nil) })
	return New(broker, emojiService)
}

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		Locale: discordgo.French, Data: data,
		Member: &discordgo.Member{User: &discordgo.User{ID: "author-id"}},
	}}
}

func TestGetName_IsAboutContractName(t *testing.T) {
	cmd := newCommand(t, &testutils.FakeBroker{}, &fakeEmojis{})
	assert.Equal(t, contract.AboutCommandName, cmd.GetName())
}

func TestGetDescriptions_UsesConfiguredDiscordIDAndLocalizedTexts(t *testing.T) {
	cmd := newCommand(t, &testutils.FakeBroker{}, &fakeEmojis{})

	descriptions := cmd.GetDescriptions(discordgo.French)

	require.Len(t, descriptions, 1)
	assert.Equal(t, "/about", descriptions[0].Name)
	assert.Equal(t, "</about:about-discord-id>", descriptions[0].CommandID)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "about.help.detailed")[0],
		descriptions[0].Description)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "about.help.tutorial")[0],
		descriptions[0].TutorialURL)
}

func TestMatches_OnlyAboutApplicationCommands(t *testing.T) {
	cmd := newCommand(t, &testutils.FakeBroker{}, &fakeEmojis{})
	cases := []struct {
		name     string
		event    *discordgo.InteractionCreate
		expected bool
	}{
		{"about slash command", interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "about"}), true},
		{"about autocomplete", interaction(discordgo.InteractionApplicationCommandAutocomplete,
			discordgo.ApplicationCommandInteractionData{Name: "about"}), true},
		{"other slash command", interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "help"}), false},
		{"message component", interaction(discordgo.InteractionMessageComponent,
			discordgo.MessageComponentInteractionData{CustomID: "about"}), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, cmd.Matches(tc.event))
		})
	}
}

func TestHandle_EmitsTraceAndEditsResponseWithAboutEmbed(t *testing.T) {
	broker := &testutils.FakeBroker{}
	emojiService := &fakeEmojis{}
	cmd := newCommand(t, broker, emojiService)
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.Handle(s, interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: "about"})))

	calls := broker.Calls()
	require.Len(t, calls, 1)
	assert.Equal(t, "Emit", calls[0].Method)
	assert.Equal(t, amqp.ExchangeRequest, calls[0].Exchange)
	assert.Equal(t, "requests.about", calls[0].RoutingKey)
	assert.Equal(t, "interaction-id", calls[0].CorrelationID)
	assert.Equal(t, amqp.RabbitMQMessage_ABOUT_REQUEST, calls[0].Message.Type)
	assert.Equal(t, "author-id", calls[0].Message.UserID)
	assert.Equal(t, amqp.Language_FR, calls[0].Message.Language)

	req := fake.Last(t)
	assert.Equal(t, "PATCH", req.Method)
	assert.Equal(t, "https://discord.com"+req.Path,
		discordgo.EndpointWebhookMessage("app", "tok", "@original"))
	payload := fake.LastEdit(t)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Contains(t, embed.Description, "<:dofusGame:1>", "game logo comes from the emoji service")
	require.Len(t, embed.Fields, 6)
	assert.Contains(t, embed.Fields[1].Name, "<:twitter:1>")
	assert.Contains(t, embed.Fields[2].Name, "<:github:1>")
	assert.Equal(t, testutils.Translations(t, discordgo.French, "about.privacy.title")[0],
		embed.Fields[4].Name)
	assert.ElementsMatch(t, []constants.EmojiMiscID{
		constants.EmojiIDGame, constants.EmojiIDTwitter, constants.EmojiIDGithub,
	}, emojiService.requested)
}

func TestHandle_BrokerFailureStillAnswersUser(t *testing.T) {
	broker := &testutils.FakeBroker{Err: errors.New("amqp down")}
	cmd := newCommand(t, broker, &fakeEmojis{})
	s, fake := testutils.NewDiscordSession(t)

	require.NoError(t, cmd.Handle(s, interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: "about"})))

	assert.Len(t, broker.Calls(), 1)
	require.Len(t, fake.Requests(), 1)
	assert.Len(t, fake.LastEdit(t).Embeds, 1)
}

func TestHandle_DiscordFailureIsSwallowed(t *testing.T) {
	cmd := newCommand(t, &testutils.FakeBroker{}, &fakeEmojis{})
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 500, Body: `{"message":"x"}`})

	assert.NoError(t, cmd.Handle(s, interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: "about"})),
		"a Discord hiccup on the answer is logged, not reported back")
	assert.Len(t, fake.Requests(), 1)
}
