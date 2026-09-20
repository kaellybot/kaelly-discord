package help

import (
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
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

type fakeEmojis struct {
	emojis.Service
}

func (fakeEmojis) GetMiscEmoji(emojiID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: string(emojiID)}
}

// fakeCommand is a DiscordCommand whose descriptions are fixed.
type fakeCommand struct {
	commands.DiscordCommand
	name         string
	descriptions []commands.Description
}

func (c fakeCommand) GetName() string { return c.name }
func (c fakeCommand) GetDescriptions(discordgo.Locale) []commands.Description {
	return c.descriptions
}

func singleDescription() fakeCommand {
	return fakeCommand{name: "item", descriptions: []commands.Description{{
		Name: "/item", CommandID: "</item:111>", Description: "shows an item", TutorialURL: "https://tuto/item.webp",
	}}}
}

func multiDescriptions() fakeCommand {
	descriptions := make([]commands.Description, 0)
	for _, sub := range []string{"get", "almanax", "rss"} {
		descriptions = append(descriptions, commands.Description{
			Name: "/config " + sub, CommandID: "</config " + sub + ":222>",
			Description: sub + " detail", TutorialURL: "https://tuto/config_" + sub + ".webp",
		})
	}
	return fakeCommand{name: "config", descriptions: descriptions}
}

type fixture struct {
	cmd      *Command
	broker   *testutils.FakeBroker
	commands *[]commands.DiscordCommand
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	viper.Set(constants.HelpID, "help-discord-id")
	t.Cleanup(func() { viper.Set(constants.HelpID, nil) })
	// Deliberately not sorted by name: the menu sorts, the select keeps order.
	cmds := []commands.DiscordCommand{
		singleDescription(), multiDescriptions(), fakeCommand{name: "about", descriptions: []commands.Description{{
			Name: "/about", CommandID: "</about:333>",
		}}},
	}
	f := &fixture{broker: &testutils.FakeBroker{}, commands: &cmds}
	f.cmd = New(f.broker, f.commands, fakeEmojis{})
	return f
}

func interaction(interactionType discordgo.InteractionType,
	data discordgo.InteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction-id", AppID: "app", Token: "tok", Type: interactionType,
		Locale: discordgo.French, Data: data,
		Member: &discordgo.Member{User: &discordgo.User{ID: authorID}},
	}}
}

func helpCommand() *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionApplicationCommand,
		discordgo.ApplicationCommandInteractionData{Name: contract.HelpCommandName})
}

func component(customID string, values ...string) *discordgo.InteractionCreate {
	return interaction(discordgo.InteractionMessageComponent,
		discordgo.MessageComponentInteractionData{CustomID: customID, Values: values})
}

type selectOption struct {
	Label   string                    `json:"label"`
	Value   string                    `json:"value"`
	Default bool                      `json:"default"`
	Emoji   *discordgo.ComponentEmoji `json:"emoji"`
}

type selectMenu struct {
	CustomID    string         `json:"custom_id"`
	Placeholder string         `json:"placeholder"`
	Options     []selectOption `json:"options"`
}

// selectMenus decodes the one-select-per-row components of a help message.
func selectMenus(t *testing.T, payload testutils.WebhookEditPayload) []selectMenu {
	t.Helper()
	menus := make([]selectMenu, 0)
	for _, raw := range payload.Components {
		var row struct {
			Components []selectMenu `json:"components"`
		}
		require.NoError(t, discordgo.Unmarshal(raw, &row))
		require.Len(t, row.Components, 1)
		menus = append(menus, row.Components[0])
	}
	return menus
}

func optionValues(menu selectMenu) ([]string, []string) {
	var values, defaults []string
	for _, option := range menu.Options {
		values = append(values, option.Value)
		if option.Default {
			defaults = append(defaults, option.Value)
		}
	}
	return values, defaults
}

func lastEdit(t *testing.T, fake *testutils.FakeDiscord) testutils.WebhookEditPayload {
	t.Helper()
	req := fake.Last(t)
	assert.Equal(t, "PATCH", req.Method)
	assert.Equal(t, "/api/v9/webhooks/app/tok/messages/@original", req.Path)
	return fake.LastEdit(t)
}

// ---- Matches / descriptions ----

func TestMatches(t *testing.T) {
	f := newFixture(t)
	cases := map[string]struct {
		event *discordgo.InteractionCreate
		want  bool
	}{
		"slash command": {helpCommand(), true},
		"other command": {interaction(discordgo.InteractionApplicationCommand,
			discordgo.ApplicationCommandInteractionData{Name: "item"}), false},
		"command select":  {component("/help", "item"), true},
		"page select":     {component("/help/config/page", "1"), true},
		"other component": {component("/helpers"), false},
		"set component":   {component("/set/1"), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, f.cmd.Matches(tc.event))
		})
	}
	assert.Equal(t, "help", f.cmd.GetName())
}

func TestGetDescriptions(t *testing.T) {
	f := newFixture(t)

	descriptions := f.cmd.GetDescriptions(discordgo.French)

	require.Len(t, descriptions, 1)
	assert.Equal(t, "/help", descriptions[0].Name)
	assert.Equal(t, "</help:help-discord-id>", descriptions[0].CommandID)
	assert.Contains(t, testutils.Translations(t, discordgo.French, "help.help.detailed"),
		descriptions[0].Description, "one of the random variants")
	assert.Equal(t, testutils.Translations(t, discordgo.French, "help.help.tutorial")[0],
		descriptions[0].TutorialURL)
}

// ---- tracing ----

func TestHandle_EveryInteractionIsTracedThroughAMQP(t *testing.T) {
	f := newFixture(t)
	s, _ := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, helpCommand())
	f.cmd.Handle(s, component("/help", "item"))

	calls := f.broker.Calls()
	require.Len(t, calls, 2)
	for _, call := range calls {
		assert.Equal(t, "Emit", call.Method)
		assert.Equal(t, amqp.ExchangeRequest, call.Exchange)
		assert.Equal(t, constants.HelpRoutingKey, call.RoutingKey)
		assert.Equal(t, "interaction-id", call.CorrelationID)
		assert.Equal(t, amqp.RabbitMQMessage_HELP_REQUEST, call.Message.Type)
		assert.Equal(t, authorID, call.Message.UserID)
		assert.Equal(t, amqp.Language_FR, call.Message.Language)
	}
}

func TestHandle_TracingFailureDoesNotBlockHelp(t *testing.T) {
	f := newFixture(t)
	f.broker.Err = errors.New("broker down")
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, helpCommand())

	require.Len(t, fake.Requests(), 1)
	assert.Len(t, lastEdit(t, fake).Embeds, 1)
}

func TestHandle_UnhandledInteractionTypesAreIgnored(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	event := helpCommand()
	event.Type = discordgo.InteractionApplicationCommandAutocomplete

	f.cmd.Handle(s, event)

	assert.Empty(t, f.broker.Calls())
	assert.Empty(t, fake.Requests())
}

// ---- menu ----

func assertMenu(t *testing.T, payload testutils.WebhookEditPayload) {
	t.Helper()
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "Liste des commandes", embed.Title)
	assert.Equal(t, constants.Color, embed.Color)
	assert.Equal(t, constants.AvatarIcon, embed.Thumbnail.URL)
	require.NotNil(t, embed.Footer)
	assert.Contains(t, embed.Footer.Text, constants.Version)
	assert.Equal(t,
		"- **about** : "+testutils.Translations(t, discordgo.French, "about.help.overview")[0]+"\n"+
			"- **config** : "+testutils.Translations(t, discordgo.French, "config.help.overview")[0]+"\n"+
			"- **item** : "+testutils.Translations(t, discordgo.French, "item.help.overview")[0]+"\n",
		embed.Description, "commands are listed by name with their overview")

	menus := selectMenus(t, payload)
	require.Len(t, menus, 1, "no page select on the menu")
	assert.Equal(t, "/help", menus[0].CustomID)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "help.commands.choices.placeholder")[0],
		menus[0].Placeholder)
	values, defaults := optionValues(menus[0])
	assert.Equal(t, []string{"menu", "item", "config", "about"}, values, "menu first, then registration order")
	assert.Equal(t, []string{"menu"}, defaults)
	assert.Equal(t, "Liste des commandes", menus[0].Options[0].Label)
	require.NotNil(t, menus[0].Options[0].Emoji)
	assert.Equal(t, string(constants.EmojiIDScroll), menus[0].Options[0].Emoji.Name)
	assert.Equal(t, "item", menus[0].Options[1].Label)
}

func TestHandle_SlashCommandShowsMenu(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, helpCommand())

	assertMenu(t, lastEdit(t, fake))
}

func TestHandle_MenuSelectionShowsMenu(t *testing.T) {
	for name, event := range map[string]*discordgo.InteractionCreate{
		"command select": component("/help", "menu"),
		"page select":    component("/help/menu/page", "3"),
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)

			f.cmd.Handle(s, event)

			assertMenu(t, lastEdit(t, fake))
		})
	}
}

func TestHandle_MenuReflectsCommandsRegisteredAfterConstruction(t *testing.T) {
	f := newFixture(t)
	*f.commands = []commands.DiscordCommand{singleDescription()}
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, helpCommand())

	payload := lastEdit(t, fake)
	values, _ := optionValues(selectMenus(t, payload)[0])
	assert.Equal(t, []string{"menu", "item"}, values)
	assert.Equal(t, 1, strings.Count(payload.Embeds[0].Description, "- **"))
}

func TestHandle_MenuWithoutCommands(t *testing.T) {
	f := newFixture(t)
	*f.commands = nil
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, helpCommand())

	payload := lastEdit(t, fake)
	assert.Equal(t, "Aucune commande disponible !", payload.Embeds[0].Description)
}

// ---- command detail ----

func TestHandle_SingleDescriptionCommandHasNoPages(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, component("/help", "item"))

	payload := lastEdit(t, fake)
	require.Len(t, payload.Embeds, 1)
	embed := payload.Embeds[0]
	assert.Equal(t, "Commande **</item:111>**", embed.Title, "the mention itself is the title")
	assert.Equal(t, "- </item:111> : shows an item\n\n\n*"+
		testutils.Translations(t, discordgo.French, "item.help.source")[0]+"*", embed.Description)
	assert.Equal(t, "https://tuto/item.webp", embed.Image.URL)
	assert.Equal(t, constants.AvatarIcon, embed.Thumbnail.URL)

	menus := selectMenus(t, payload)
	require.Len(t, menus, 1)
	_, defaults := optionValues(menus[0])
	assert.Equal(t, []string{"item"}, defaults, "the displayed command is preselected")
}

func TestHandle_MultiDescriptionCommandShowsFirstPageAndPageSelect(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, component("/help", "config"))

	payload := lastEdit(t, fake)
	embed := payload.Embeds[0]
	assert.Equal(t, "Commande **config** : </config get:222>", embed.Title)
	assert.True(t, strings.HasPrefix(embed.Description, "- </config get:222> : get detail\n"), embed.Description)
	assert.Equal(t, "https://tuto/config_get.webp", embed.Image.URL)

	menus := selectMenus(t, payload)
	require.Len(t, menus, 2)
	_, defaults := optionValues(menus[0])
	assert.Equal(t, []string{"config"}, defaults)

	pages := menus[1]
	assert.Equal(t, contract.CraftHelpPageCustomID("config"), pages.CustomID)
	assert.Equal(t, testutils.Translations(t, discordgo.French, "help.commands.pages.placeholder")[0],
		pages.Placeholder)
	values, pageDefaults := optionValues(pages)
	assert.Equal(t, []string{"0", "1", "2"}, values)
	assert.Equal(t, []string{"0"}, pageDefaults)
	assert.Equal(t, "/config almanax", pages.Options[1].Label)
}

func TestHandle_PageSelectShowsRequestedSubCommand(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	f.cmd.Handle(s, component("/help/config/page", "2"))

	payload := lastEdit(t, fake)
	embed := payload.Embeds[0]
	assert.Equal(t, "Commande **config** : </config rss:222>", embed.Title)
	assert.Equal(t, "https://tuto/config_rss.webp", embed.Image.URL)
	_, pageDefaults := optionValues(selectMenus(t, payload)[1])
	assert.Equal(t, []string{"2"}, pageDefaults)
}

func TestHandle_DetailIsLocalized(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	event := component("/help", "item")
	event.Locale = discordgo.EnglishGB

	f.cmd.Handle(s, event)

	embed := lastEdit(t, fake).Embeds[0]
	assert.Equal(t, "**</item:111>** command", embed.Title)
	assert.Contains(t, embed.Description, testutils.Translations(t, discordgo.EnglishGB, "item.help.source")[0])
}

func TestHandle_InvalidSelectionsPanic(t *testing.T) {
	cases := map[string]struct {
		event   *discordgo.InteractionCreate
		wantErr string
	}{
		"no value":          {component("/help"), commands.ErrInvalidInteraction.Error()},
		"two values":        {component("/help", "item", "config"), commands.ErrInvalidInteraction.Error()},
		"page not a number": {component("/help/config/page", "two"), commands.ErrInvalidInteraction.Error()},
		"foreign custom ID": {component("/set/1", "item"), commands.ErrInvalidInteraction.Error()},
		"unknown command": {component("/help", "ghost"),
			"cannot find command (ghost) while searching for its help description, panicking"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			s, fake := testutils.NewDiscordSession(t)

			assert.PanicsWithError(t, tc.wantErr, func() { f.cmd.Handle(s, tc.event) })
			assert.Empty(t, fake.Requests())
			assert.Len(t, f.broker.Calls(), 1, "tracing happens before validation")
		})
	}
}

func TestHandle_OutOfRangePagePanics(t *testing.T) {
	// A page index beyond the command's descriptions is not validated and ends
	// in a runtime index-out-of-range panic (recovered upstream by HandlePanic).
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)

	assert.Panics(t, func() { f.cmd.Handle(s, component("/help/config/page", "3")) })
	assert.Panics(t, func() { f.cmd.Handle(s, component("/help/config/page", "-1")) })
	assert.Empty(t, fake.Requests())
}

func TestHandle_DiscordFailureIsSwallowed(t *testing.T) {
	f := newFixture(t)
	s, fake := testutils.NewDiscordSession(t)
	fake.Respond(testutils.DiscordResponse{PathContains: "/webhooks/", Status: 400, Body: `{}`})

	assert.NotPanics(t, func() {
		f.cmd.Handle(s, helpCommand())
		f.cmd.Handle(s, component("/help", "item"))
	})
	assert.Len(t, fake.Requests(), 2)
}
