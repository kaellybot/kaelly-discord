package discord

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeEmojiService struct {
	emojis.Service
}

func (fakeEmojiService) GetMiscEmoji(emojiID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: string(emojiID)}
}

func pageID(page int) string { return fmt.Sprintf("/page/%d", page) }

type expectedButton struct {
	customID string
	label    string
	emoji    constants.EmojiMiscID
	disabled bool
}

func toExpected(t *testing.T, components []discordgo.MessageComponent) []expectedButton {
	t.Helper()
	buttons := make([]expectedButton, 0, len(components))
	for _, component := range components {
		button, ok := component.(discordgo.Button)
		require.True(t, ok, "%T is not a button", component)
		assert.Equal(t, discordgo.SecondaryButton, button.Style)
		require.NotNil(t, button.Emoji)
		buttons = append(buttons, expectedButton{
			customID: button.CustomID, label: button.Label,
			emoji: constants.EmojiMiscID(button.Emoji.Name), disabled: button.Disabled,
		})
	}
	return buttons
}

func TestGetPaginationButtons_NoButtonsForASinglePage(t *testing.T) {
	for _, pages := range []int{0, 1} {
		buttons := GetPaginationButtons(0, pages, pageID, discordgo.EnglishGB, fakeEmojiService{})
		assert.NotNil(t, buttons)
		assert.Empty(t, buttons, "%d pages", pages)
	}
}

func TestGetPaginationButtons_Navigation(t *testing.T) {
	first := func(disabled bool) expectedButton {
		return expectedButton{"/page/0", "First page", constants.EmojiIDFirst, disabled}
	}
	last := func(page int, disabled bool) expectedButton {
		return expectedButton{pageID(page), "Last page", constants.EmojiIDLast, disabled}
	}
	previous := func(page int) expectedButton {
		return expectedButton{pageID(page), "Previous page", constants.EmojiIDPrevious, false}
	}
	next := func(page int) expectedButton {
		return expectedButton{pageID(page), "Next page", constants.EmojiIDNext, false}
	}

	cases := []struct {
		name        string
		page, pages int
		expected    []expectedButton
	}{
		{"two pages, on first", 0, 2, []expectedButton{first(true), last(1, false)}},
		{"two pages, on last", 1, 2, []expectedButton{first(false), last(1, true)}},
		{"on first of many", 0, 5, []expectedButton{first(true), next(1)}},
		{"second page goes back to first rather than previous", 1, 5, []expectedButton{first(false), next(2)}},
		{"middle page", 2, 5, []expectedButton{previous(1), next(3)}},
		{"before last page jumps to last rather than next", 3, 5, []expectedButton{previous(2), last(4, false)}},
		{"on last of many", 4, 5, []expectedButton{previous(3), last(4, true)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buttons := GetPaginationButtons(tc.page, tc.pages, pageID, discordgo.EnglishGB, fakeEmojiService{})
			assert.Equal(t, tc.expected, toExpected(t, buttons))
		})
	}
}

func TestGetPaginationButtons_OutOfRangePageIsClamped(t *testing.T) {
	// A stale button (the list shrank since it was sent) must never offer a
	// page outside of [0, pages-1].
	beyond := GetPaginationButtons(10, 5, pageID, discordgo.EnglishGB, fakeEmojiService{})
	assert.Equal(t, []expectedButton{
		{"/page/3", "Previous page", constants.EmojiIDPrevious, false},
		{"/page/4", "Last page", constants.EmojiIDLast, true},
	}, toExpected(t, beyond), "clamped on the last page")

	negative := GetPaginationButtons(-3, 5, pageID, discordgo.EnglishGB, fakeEmojiService{})
	assert.Equal(t, []expectedButton{
		{"/page/0", "First page", constants.EmojiIDFirst, true},
		{"/page/1", "Next page", constants.EmojiIDNext, false},
	}, toExpected(t, negative), "clamped on the first page")

	twoPages := GetPaginationButtons(7, 2, pageID, discordgo.EnglishGB, fakeEmojiService{})
	assert.Equal(t, []expectedButton{
		{"/page/0", "First page", constants.EmojiIDFirst, false},
		{"/page/1", "Last page", constants.EmojiIDLast, true},
	}, toExpected(t, twoPages))
}

func TestGetPaginationButtons_LabelsAreLocalized(t *testing.T) {
	buttons := GetPaginationButtons(2, 5, pageID, discordgo.French, fakeEmojiService{})

	got := toExpected(t, buttons)
	require.Len(t, got, 2)
	assert.Equal(t, "Page précédente", got[0].label)
	assert.Equal(t, "Page suivante", got[1].label)
}

func TestSliceFields_OneFieldPerChunkWithIndex(t *testing.T) {
	type call struct {
		index int
		items []string
	}
	var calls []call
	fields := SliceFields([]string{"a", "b", "c", "d", "e"}, 2,
		func(i int, items []string) *discordgo.MessageEmbedField {
			calls = append(calls, call{i, items})
			return &discordgo.MessageEmbedField{Name: fmt.Sprint(i), Value: fmt.Sprint(items)}
		})

	assert.Equal(t, []call{{0, []string{"a", "b"}}, {1, []string{"c", "d"}}, {2, []string{"e"}}}, calls)
	require.Len(t, fields, 3)
	assert.Equal(t, "[e]", fields[2].Value)
}

func TestSliceFields_NoItemsNoFields(t *testing.T) {
	fields := SliceFields([]int{}, 3, func(int, []int) *discordgo.MessageEmbedField {
		t.Fatal("must not be called")
		return nil
	})

	assert.NotNil(t, fields)
	assert.Empty(t, fields)
}

func TestSliceButtons_AtMostFivePerRow(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7}
	rows := SliceButtons(items, func(item int) discordgo.Button {
		return discordgo.Button{CustomID: fmt.Sprint(item)}
	})

	require.Len(t, rows, 2)
	ids := make([][]string, 0, len(rows))
	for _, row := range rows {
		rowIDs := make([]string, 0)
		for _, component := range row.Components {
			button, ok := component.(discordgo.Button)
			require.True(t, ok)
			rowIDs = append(rowIDs, button.CustomID)
		}
		ids = append(ids, rowIDs)
	}
	assert.Equal(t, [][]string{{"1", "2", "3", "4", "5"}, {"6", "7"}}, ids)
}

func TestSliceButtons_NoItemsNoRows(t *testing.T) {
	rows := SliceButtons([]int{}, func(int) discordgo.Button { return discordgo.Button{} })

	assert.NotNil(t, rows)
	assert.Empty(t, rows, "an empty action row would be rejected by Discord")
}

func TestGhostInlineField_IsInvisibleAndFullWidth(t *testing.T) {
	field := GhostInlineField()

	assert.Equal(t, "\u200b", field.Name)
	assert.Empty(t, field.Value)
	assert.False(t, field.Inline)
	assert.NotSame(t, field, GhostInlineField(), "callers may mutate their own copy")
}

func TestBuildDefaultFooter_Localized(t *testing.T) {
	cases := []struct {
		locale   discordgo.Locale
		expected string
	}{
		{discordgo.EnglishGB, "KaellyBot v2.0.0 • Slash commands available!"},
		{discordgo.French, "KaellyBot v2.0.0 • Commandes Slash disponibles !"},
	}
	for _, tc := range cases {
		footer := BuildDefaultFooter(tc.locale)
		assert.Equal(t, tc.expected, footer.Text, tc.locale)
		assert.Equal(t, constants.AvatarIcon, footer.IconURL)
	}
}
