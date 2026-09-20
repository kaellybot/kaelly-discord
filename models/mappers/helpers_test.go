package mappers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/equipments"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/twitters"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

// fakeEmojis renders every emoji as a readable marker so tests can check
// which emoji ended up where.
type fakeEmojis struct{ emojis.Service }

func (fakeEmojis) GetMiscStringEmoji(emojiID constants.EmojiMiscID) string {
	return "<misc:" + string(emojiID) + ">"
}

func (fakeEmojis) GetMiscEmoji(emojiID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: "misc:" + string(emojiID)}
}

func (fakeEmojis) GetEntityStringEmoji(entityID string, entityType constants.EmojiEntity) string {
	return "<" + string(entityType) + ":" + entityID + ">"
}

func (fakeEmojis) GetEntityEmoji(entityID string, entityType constants.EmojiEntity) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: string(entityType) + ":" + entityID}
}

func (fakeEmojis) GetEquipmentStringEmoji(equipmentType amqp.EquipmentType) string {
	return "<equip:" + equipmentType.String() + ">"
}

func (fakeEmojis) GetEquipmentEmoji(equipmentType amqp.EquipmentType) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: "equip:" + equipmentType.String()}
}

func (fakeEmojis) GetItemTypeStringEmoji(itemType amqp.ItemType) string {
	return "<item:" + itemType.String() + ">"
}

func (fakeEmojis) GetItemTypeEmoji(itemType amqp.ItemType) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: "item:" + itemType.String()}
}

func (fakeEmojis) GetSetBonusEmoji(equipedItemNumber int) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: fmt.Sprintf("bonus:%d", equipedItemNumber)}
}

type fakeServers struct {
	servers.Service
	servers []entities.Server
}

func (f fakeServers) GetServer(id string) (entities.Server, bool) {
	for _, server := range f.servers {
		if server.ID == id {
			return server, true
		}
	}
	return entities.Server{}, false
}

type fakeBooks struct {
	books.Service
	jobs   []entities.Job
	cities []entities.City
	orders []entities.Order
}

func (f fakeBooks) GetJob(id string) (entities.Job, bool) {
	for _, job := range f.jobs {
		if job.ID == id {
			return job, true
		}
	}
	return entities.Job{}, false
}

func (f fakeBooks) GetJobs() []entities.Job { return f.jobs }

func (f fakeBooks) GetCity(id string) (entities.City, bool) {
	for _, city := range f.cities {
		if city.ID == id {
			return city, true
		}
	}
	return entities.City{}, false
}

func (f fakeBooks) GetCities() []entities.City { return f.cities }

func (f fakeBooks) GetOrder(id string) (entities.Order, bool) {
	for _, order := range f.orders {
		if order.ID == id {
			return order, true
		}
	}
	return entities.Order{}, false
}

func (f fakeBooks) GetOrders() []entities.Order { return f.orders }

type fakeFeeds struct {
	feeds.Service
	feedTypes []entities.FeedType
}

func (f fakeFeeds) GetFeedType(id string) *entities.FeedType {
	for i := range f.feedTypes {
		if f.feedTypes[i].ID == id {
			return &f.feedTypes[i]
		}
	}
	return nil
}

type fakeTwitters struct {
	twitters.Service
	accounts []entities.TwitterAccount
}

func (f fakeTwitters) GetTwitterAccount(id string) *entities.TwitterAccount {
	for i := range f.accounts {
		if f.accounts[i].ID == id {
			return &f.accounts[i]
		}
	}
	return nil
}

type fakeCharacteristics struct {
	characteristics.Service
	characs map[string]entities.Characteristic
}

func (f fakeCharacteristics) GetCharacteristic(id string) (entities.Characteristic, bool) {
	charac, found := f.characs[id]
	return charac, found
}

type fakeEquipments struct {
	equipments.Service
	effects map[string]entities.WeaponAreaEffect
}

func (f fakeEquipments) GetWeaponAreaEffect(id string) (entities.WeaponAreaEffect, bool) {
	effect, found := f.effects[id]
	return effect, found
}

func testServer() entities.Server {
	return entities.Server{ID: "srv", Icon: "srv.png", Labels: []entities.ServerLabel{
		{ServerID: "srv", Locale: amqp.Language_EN, Label: "Imagiro"},
		{ServerID: "srv", Locale: amqp.Language_FR, Label: "Imagiro-FR"},
	}}
}

// assertRendered fails when a translation came back as its raw key or as an
// unrendered template (di18n returns the raw text on template errors).
func assertRendered(t *testing.T, text string) {
	t.Helper()
	assert.NotEmpty(t, text)
	assert.NotContains(t, text, "{{", "template was not rendered: %q", text)
}

// assertTranslated checks text is one of the (placeholder-free) variants of key.
func assertTranslated(t *testing.T, lg discordgo.Locale, key, text string) {
	t.Helper()
	assert.Contains(t, testutils.Translations(t, lg, key), text)
}

func actionRows(t *testing.T, components *[]discordgo.MessageComponent) []discordgo.ActionsRow {
	t.Helper()
	require.NotNil(t, components)
	rows := make([]discordgo.ActionsRow, 0, len(*components))
	for _, component := range *components {
		row, ok := component.(discordgo.ActionsRow)
		require.True(t, ok, "expected an actions row, got %T", component)
		rows = append(rows, row)
	}
	return rows
}

func buttons(t *testing.T, row discordgo.ActionsRow) []discordgo.Button {
	t.Helper()
	result := make([]discordgo.Button, 0, len(row.Components))
	for _, component := range row.Components {
		button, ok := component.(discordgo.Button)
		require.True(t, ok, "expected a button, got %T", component)
		result = append(result, button)
	}
	return result
}

func selectMenu(t *testing.T, row discordgo.ActionsRow) discordgo.SelectMenu {
	t.Helper()
	require.Len(t, row.Components, 1)
	menu, ok := row.Components[0].(discordgo.SelectMenu)
	require.True(t, ok, "expected a select menu, got %T", row.Components[0])
	assert.Equal(t, discordgo.StringSelectMenu, menu.MenuType)
	return menu
}

func singleEmbed(t *testing.T, embeds *[]*discordgo.MessageEmbed) *discordgo.MessageEmbed {
	t.Helper()
	require.NotNil(t, embeds)
	require.Len(t, *embeds, 1)
	return (*embeds)[0]
}

func assertDefaultFooter(t *testing.T, footer *discordgo.MessageEmbedFooter) {
	t.Helper()
	require.NotNil(t, footer)
	assert.Contains(t, footer.Text, constants.Name)
	assert.Contains(t, footer.Text, constants.Version)
	assertRendered(t, footer.Text)
	assert.Equal(t, constants.AvatarIcon, footer.IconURL)
}

// assertInOrder checks each fragment appears in text, in the given order.
func assertInOrder(t *testing.T, text string, fragments ...string) {
	t.Helper()
	last := -1
	for _, fragment := range fragments {
		index := strings.Index(text, fragment)
		require.GreaterOrEqual(t, index, 0, "%q not found in %q", fragment, text)
		assert.Greater(t, index, last, "%q is out of order in %q", fragment, text)
		last = index
	}
}

func source() *amqp.Source {
	return &amqp.Source{Name: "dofusdude", Icon: "src.png", Url: "https://src"}
}

func assertSourceAuthor(t *testing.T, author *discordgo.MessageEmbedAuthor) {
	t.Helper()
	require.NotNil(t, author)
	assert.Equal(t, discordgo.MessageEmbedAuthor{Name: "dofusdude", URL: "https://src", IconURL: "src.png"}, *author)
}
