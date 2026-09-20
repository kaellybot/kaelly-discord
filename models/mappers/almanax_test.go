package mappers

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func day(year int, month time.Month, dayOfMonth int) time.Time {
	return time.Date(year, month, dayOfMonth, 0, 0, 0, 0, time.UTC)
}

func almanaxOf(date time.Time, bonus string) *amqp.Almanax {
	return &amqp.Almanax{
		Bonus:  bonus,
		Reward: 1234567,
		Date:   timestamppb.New(date),
		Tribute: &amqp.Almanax_Tribute{
			Quantity: 3,
			Item:     &amqp.Almanax_Tribute_Item{Name: "Wheat", Icon: "wheat.png", Type: amqp.ItemType_RESOURCE_TYPE},
		},
	}
}

func TestMapAlmanaxRequests_CopyFields(t *testing.T) {
	date := day(2024, time.July, 14)
	request := MapAlmanaxRequest(&date, "author", discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_REQUEST, request.Type)
	assert.Equal(t, amqp.Language_FR, request.Language)
	assert.Equal(t, "author", request.UserID)
	require.NotNil(t, request.EncyclopediaAlmanaxRequest)
	assert.True(t, date.Equal(request.EncyclopediaAlmanaxRequest.Date.AsTime()))

	resource := MapAlmanaxResourceRequest(15, "author", discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_REQUEST, resource.Type)
	require.NotNil(t, resource.EncyclopediaAlmanaxResourceRequest)
	assert.Equal(t, int64(15), resource.EncyclopediaAlmanaxResourceRequest.Duration)

	list := MapAlmanaxEffectListRequest("fish", "author", discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, list.Type)
	require.NotNil(t, list.EncyclopediaListRequest)
	assert.Equal(t, "fish", list.EncyclopediaListRequest.Query)
	assert.Equal(t, amqp.EncyclopediaListRequest_ALMANAX_EFFECT, list.EncyclopediaListRequest.Type)
}

func TestMapAlmanaxEffectRequest_QueryOrDate(t *testing.T) {
	query := "fish"
	date := day(2024, time.July, 14)

	byQuery := MapAlmanaxEffectRequest(&query, nil, 2, "author", discordgo.EnglishGB)
	require.NotNil(t, byQuery)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_REQUEST, byQuery.Type)
	assert.Equal(t, "author", byQuery.UserID)
	effect := byQuery.EncyclopediaAlmanaxEffectRequest
	require.NotNil(t, effect)
	assert.Equal(t, amqp.EncyclopediaAlmanaxEffectRequest_QUERY, effect.Type)
	assert.Equal(t, "fish", effect.Query)
	assert.Nil(t, effect.Date)
	assert.Equal(t, int64(2*constants.MaxAlmanaxEffectPerEmbed), effect.Offset)
	assert.Equal(t, int64(constants.MaxAlmanaxEffectPerEmbed), effect.Size)

	byDate := MapAlmanaxEffectRequest(nil, &date, 0, "author", discordgo.EnglishGB)
	require.NotNil(t, byDate)
	effect = byDate.EncyclopediaAlmanaxEffectRequest
	assert.Equal(t, amqp.EncyclopediaAlmanaxEffectRequest_DATE, effect.Type)
	assert.Empty(t, effect.Query)
	assert.True(t, date.Equal(effect.Date.AsTime()))
	assert.Zero(t, effect.Offset)

	both := MapAlmanaxEffectRequest(&query, &date, 0, "author", discordgo.EnglishGB)
	assert.Equal(t, amqp.EncyclopediaAlmanaxEffectRequest_QUERY, both.EncyclopediaAlmanaxEffectRequest.Type,
		"query takes precedence over date")
	assert.Nil(t, both.EncyclopediaAlmanaxEffectRequest.Date)

	assert.Nil(t, MapAlmanaxEffectRequest(nil, nil, 0, "author", discordgo.EnglishGB))
}

func TestMapAlmanaxToWebhook_MissingAlmanaxGivesContentOnly(t *testing.T) {
	edit := MapAlmanaxToWebhook(&amqp.EncyclopediaAlmanaxAnswer{Source: source()}, discordgo.French, fakeEmojis{})

	require.NotNil(t, edit.Content)
	assertTranslated(t, discordgo.French, "almanax.day.missing", *edit.Content)
	assert.Nil(t, edit.Embeds)
	assert.Nil(t, edit.Components)
}

func TestMapAlmanaxToWebhook_DayEmbed(t *testing.T) {
	date := day(2024, time.July, 14)
	answer := &amqp.EncyclopediaAlmanaxAnswer{Almanax: almanaxOf(date, "More fish"), Source: source()}

	edit := MapAlmanaxToWebhook(answer, discordgo.EnglishGB, fakeEmojis{})

	assert.Nil(t, edit.Content)
	embed := singleEmbed(t, edit.Embeds)
	assert.Contains(t, embed.Title, "July 14, 2024")
	assertRendered(t, embed.Title)
	assert.Equal(t, "https://www.krosmoz.com/en/almanax/2024-07-14", embed.URL)
	summer := constants.GetSeason(date)
	assert.Equal(t, "Summer", summer.Name)
	assert.Equal(t, summer.Color, embed.Color)
	assert.Equal(t, summer.AlmanaxIcon, embed.Thumbnail.URL)
	assert.Equal(t, "wheat.png", embed.Image.URL)
	assertSourceAuthor(t, embed.Author)
	assertDefaultFooter(t, embed.Footer)

	require.Len(t, embed.Fields, 3)
	assertTranslated(t, discordgo.EnglishGB, "almanax.day.bonus.title", embed.Fields[0].Name)
	assert.Equal(t, "More fish", embed.Fields[0].Value)
	assertTranslated(t, discordgo.EnglishGB, "almanax.day.tribute.title", embed.Fields[1].Name)
	assert.Contains(t, embed.Fields[1].Value, "<item:RESOURCE_TYPE>")
	assert.Contains(t, embed.Fields[1].Value, "**3**x Wheat")
	assertTranslated(t, discordgo.EnglishGB, "almanax.day.reward.title", embed.Fields[2].Name)
	assert.Contains(t, embed.Fields[2].Value, "**1,234,567**", "reward is formatted for the locale")
	assert.Contains(t, embed.Fields[2].Value, "<misc:kama>")

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 1)
	btns := buttons(t, rows[0])
	require.Len(t, btns, 3)
	previous, ok := contract.ExtractAlmanaxDayCustomID(btns[0].CustomID)
	require.True(t, ok, btns[0].CustomID)
	assert.True(t, day(2024, time.July, 13).Equal(*previous))
	assertTranslated(t, discordgo.EnglishGB, "almanax.day.previous", btns[0].Label)
	assert.Equal(t, "misc:previous", btns[0].Emoji.Name)
	next, ok := contract.ExtractAlmanaxDayCustomID(btns[1].CustomID)
	require.True(t, ok, btns[1].CustomID)
	assert.True(t, day(2024, time.July, 15).Equal(*next))
	assertTranslated(t, discordgo.EnglishGB, "almanax.day.next", btns[1].Label)
	assert.Equal(t, "misc:next", btns[1].Emoji.Name)
	effectDate, page, ok := contract.ExtractAlmanaxEffectCustomID(btns[2].CustomID)
	require.True(t, ok, btns[2].CustomID)
	assert.True(t, date.Equal(*effectDate))
	assert.Equal(t, constants.DefaultPage, page)
	assert.Equal(t, discordgo.PrimaryButton, btns[2].Style)
	assert.Equal(t, discordgo.SecondaryButton, btns[0].Style)
}

func TestMapAlmanaxToWebhook_DayEmbedIsLocalized(t *testing.T) {
	date := day(2024, time.January, 1)
	answer := &amqp.EncyclopediaAlmanaxAnswer{Almanax: almanaxOf(date, "Bonus"), Source: source()}

	edit := MapAlmanaxToWebhook(answer, discordgo.French, fakeEmojis{})

	embed := singleEmbed(t, edit.Embeds)
	assert.Contains(t, embed.Title, "janvier")
	assert.Contains(t, embed.URL, "2024-01-01")
	assert.NotContains(t, embed.Fields[2].Value, "1,234,567")
	assert.Equal(t, constants.GetSeason(date).Color, embed.Color)
	assertTranslated(t, discordgo.French, "almanax.day.next", buttons(t, actionRows(t, edit.Components)[0])[1].Label)
}

func TestMapAlmanaxEffectsToWebhook_UnknownEffectIsQueryMismatch(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxEffectAnswer{Query: "fsh"}

	edit := MapAlmanaxEffectsToWebhook(answer, discordgo.EnglishGB, fakeEmojis{})

	assert.Equal(t, MapQueryMismatch("fsh", discordgo.EnglishGB).Embeds, edit.Embeds)
	require.NotNil(t, edit.Content)
	assert.Contains(t, *edit.Content, "fsh")
	assert.Nil(t, edit.Components)
}

func TestMapAlmanaxEffectsToWebhook_KnownEffectWithoutDays(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxEffectAnswer{Query: "fish", EffectName: "Fish"}

	edit := MapAlmanaxEffectsToWebhook(answer, discordgo.French, fakeEmojis{})

	require.NotNil(t, edit.Content)
	assertTranslated(t, discordgo.French, "almanax.effect.missing", *edit.Content)
	assert.Nil(t, edit.Embeds)
	assert.Nil(t, edit.Components)
}

func TestMapAlmanaxEffectsToWebhook_ListsDaysWithPaginationAndChoices(t *testing.T) {
	first, second := day(2024, time.March, 3), day(2024, time.April, 4)
	answer := &amqp.EncyclopediaAlmanaxEffectAnswer{
		Query: "fish", EffectName: "Fish", Page: 1, Pages: 4, Total: 18, Source: source(),
		Almanaxes: []*amqp.Almanax{almanaxOf(first, "Bonus A"), almanaxOf(second, "Bonus B")},
	}

	edit := MapAlmanaxEffectsToWebhook(answer, discordgo.EnglishGB, fakeEmojis{})

	embed := singleEmbed(t, edit.Embeds)
	assert.Equal(t, "Fish", embed.Title, "the resolved effect name is shown, not the raw query")
	assertRendered(t, embed.Description)
	assert.Contains(t, embed.Description, "18")
	assert.Contains(t, embed.Description, "page 2 on 4")
	assert.Equal(t, constants.GetUnknownSeason().AlmanaxIcon, embed.Thumbnail.URL)
	assert.Equal(t, constants.Color, embed.Color)
	assertSourceAuthor(t, embed.Author)
	assertDefaultFooter(t, embed.Footer)
	require.Len(t, embed.Fields, 2)
	for i, date := range []time.Time{first, second} {
		assert.Contains(t, embed.Fields[i].Name, fmt.Sprintf("<t:%d:D>", date.Unix()))
		assert.Contains(t, embed.Fields[i].Name, "<misc:calendar>")
		assert.False(t, embed.Fields[i].Inline)
	}
	assert.Equal(t, "Bonus A", embed.Fields[0].Value)
	assert.Equal(t, "Bonus B", embed.Fields[1].Value)

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 2)
	pagination := buttons(t, rows[0])
	require.Len(t, pagination, 2)
	for i, expectedPage := range []int{0, 2} {
		date, page, ok := contract.ExtractAlmanaxEffectCustomID(pagination[i].CustomID)
		require.True(t, ok, pagination[i].CustomID)
		assert.True(t, first.Equal(*date), "the first day carries the effect across pages")
		assert.Equal(t, expectedPage, page)
	}

	menu := selectMenu(t, rows[1])
	assert.Equal(t, contract.CraftAlmanaxDayChoiceCustomID(), menu.CustomID)
	assertTranslated(t, discordgo.EnglishGB, "almanax.effect.choice.placeholder", menu.Placeholder)
	require.Len(t, menu.Options, 2)
	assert.Equal(t, strconv.FormatInt(first.Unix(), 10), menu.Options[0].Value)
	assert.Contains(t, menu.Options[0].Label, "03/03/2024")
	assert.Equal(t, strconv.FormatInt(second.Unix(), 10), menu.Options[1].Value)
	assert.Contains(t, menu.Options[1].Label, "04/04/2024")
	assert.Equal(t, "misc:calendar", menu.Options[1].Emoji.Name)
}

func TestMapAlmanaxEffectsToWebhook_SinglePageHasNoPagination(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxEffectAnswer{
		Query: "fish", EffectName: "Fish", Pages: 1, Total: 1,
		Almanaxes: []*amqp.Almanax{almanaxOf(day(2024, time.March, 3), "Bonus")},
	}

	edit := MapAlmanaxEffectsToWebhook(answer, discordgo.EnglishGB, fakeEmojis{})

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 1)
	selectMenu(t, rows[0])
}

func TestMapAlmanaxResourceToWebhook_SortsTributesAndMultipliesByCharacters(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxResourceAnswer{
		Duration: 10,
		Source:   source(),
		Tributes: []*amqp.EncyclopediaAlmanaxResourceAnswer_Tribute{
			{ItemName: "Zinc", ItemType: amqp.ItemType_RESOURCE_TYPE, Quantity: 1},
			{ItemName: "Écorce", ItemType: amqp.ItemType_RESOURCE_TYPE, Quantity: 2},
			{ItemName: "Potion", ItemType: amqp.ItemType_CONSUMABLE_TYPE, Quantity: 3},
			{ItemName: "abeille", ItemType: amqp.ItemType_RESOURCE_TYPE, Quantity: 4},
		},
	}

	before := time.Now().UTC()
	edit := MapAlmanaxResourceToWebhook(answer, 3, discordgo.French, fakeEmojis{})
	after := time.Now().UTC()

	embed := singleEmbed(t, edit.Embeds)
	assertRendered(t, embed.Description)
	// Item type first, then locale-aware name order (accents and case do not push items last).
	assertInOrder(t, embed.Description,
		"<item:CONSUMABLE_TYPE> **9**x Potion",
		"<item:RESOURCE_TYPE> **12**x abeille",
		"<item:RESOURCE_TYPE> **6**x Écorce",
		"<item:RESOURCE_TYPE> **3**x Zinc",
	)
	assertRendered(t, embed.Title)
	matched := false
	for _, now := range []time.Time{before, after} {
		start := day(now.Year(), now.Month(), now.Day())
		end := start.AddDate(0, 0, 10)
		matched = matched || strings.Contains(embed.Title, fmt.Sprintf("<t:%d:D>", start.Unix())) &&
			strings.Contains(embed.Title, fmt.Sprintf("<t:%d:D>", end.Unix()))
	}
	assert.True(t, matched, "title spans today (UTC) to today + duration: %q", embed.Title)
	assert.Equal(t, constants.GetUnknownSeason().AlmanaxIcon, embed.Thumbnail.URL)
	assertSourceAuthor(t, embed.Author)
	assertDefaultFooter(t, embed.Footer)

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 2)
	durationMenu := selectMenu(t, rows[0])
	characters, ok := contract.ExtractAlmanaxResourceDurationCustomID(durationMenu.CustomID)
	require.True(t, ok, durationMenu.CustomID)
	assert.Equal(t, int64(3), characters, "duration menu keeps the character number")
	assertTranslated(t, discordgo.French, "almanax.resource.duration.placeholder", durationMenu.Placeholder)
	assertOptions(t, durationMenu.Options, constants.GetAlmanaxDayDuration(), "10")

	characterMenu := selectMenu(t, rows[1])
	duration, ok := contract.ExtractAlmanaxResourceCharacterCustomID(characterMenu.CustomID)
	require.True(t, ok, characterMenu.CustomID)
	assert.Equal(t, int64(10), duration, "character menu keeps the duration")
	assertTranslated(t, discordgo.French, "almanax.resource.character.placeholder", characterMenu.Placeholder)
	assertOptions(t, characterMenu.Options, constants.GetCharacterNumbers(), "3")
}

func TestMapAlmanaxResourceToWebhook_NoTributesAndUnlistedDuration(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxResourceAnswer{Duration: 3, Source: source()}

	edit := MapAlmanaxResourceToWebhook(answer, 1, discordgo.EnglishGB, fakeEmojis{})

	embed := singleEmbed(t, edit.Embeds)
	assertRendered(t, embed.Description)
	assert.NotContains(t, embed.Description, "<item:")
	rows := actionRows(t, edit.Components)
	assertOptions(t, selectMenu(t, rows[0]).Options, constants.GetAlmanaxDayDuration(), "7")
}

func TestMapAlmanaxResourceToWebhook_UnlistedDurationSelectsTheClosestOption(t *testing.T) {
	for duration, expected := range map[int64]string{1: "7", 6: "7", 12: "12", 99: "30"} {
		t.Run(strconv.FormatInt(duration, 10), func(t *testing.T) {
			answer := &amqp.EncyclopediaAlmanaxResourceAnswer{Duration: duration, Source: source()}

			edit := MapAlmanaxResourceToWebhook(answer, 1, discordgo.EnglishGB, fakeEmojis{})

			options := selectMenu(t, actionRows(t, edit.Components)[0]).Options
			assertOptions(t, options, constants.GetAlmanaxDayDuration(), expected)
		})
	}
}

func TestMapAlmanaxResourceToWebhook_LeavesCallerTributesUnsorted(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxResourceAnswer{
		Duration: 7, Source: source(),
		Tributes: []*amqp.EncyclopediaAlmanaxResourceAnswer_Tribute{
			{ItemName: "Zinc", ItemType: amqp.ItemType_RESOURCE_TYPE, Quantity: 1},
			{ItemName: "Potion", ItemType: amqp.ItemType_CONSUMABLE_TYPE, Quantity: 3},
		},
	}

	MapAlmanaxResourceToWebhook(answer, 1, discordgo.EnglishGB, fakeEmojis{})

	names := make([]string, 0, len(answer.Tributes))
	for _, tribute := range answer.Tributes {
		names = append(names, tribute.ItemName)
	}
	assert.Equal(t, []string{"Zinc", "Potion"}, names, "the caller's slice is left untouched")
}

func TestMapAlmanaxToWebhook_MissingTributeAndSourceDoNotPanic(t *testing.T) {
	almanax := almanaxOf(day(2024, time.July, 14), "More fish")
	almanax.Tribute = nil
	answer := &amqp.EncyclopediaAlmanaxAnswer{Almanax: almanax}

	embed := singleEmbed(t, MapAlmanaxToWebhook(answer, discordgo.EnglishGB, fakeEmojis{}).Embeds)

	assert.Empty(t, embed.Image.URL)
	assert.Equal(t, &discordgo.MessageEmbedAuthor{}, embed.Author)
	assertRendered(t, embed.Fields[1].Value)
}

func TestMapAlmanaxEffectsToEmbeds_TitleFallsBackToQueryWithoutEffectName(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxEffectAnswer{
		Query: "fish", Pages: 1, Total: 1,
		Almanaxes: []*amqp.Almanax{almanaxOf(day(2024, time.March, 3), "Bonus")},
	}

	embed := singleEmbed(t, mapAlmanaxEffectsToEmbeds(answer, discordgo.EnglishGB, fakeEmojis{}))

	assert.Equal(t, "fish", embed.Title)
}

func TestMapAlmanaxEffectsToWebhook_MissingSourceLeavesAuthorEmpty(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxEffectAnswer{
		Query: "fish", EffectName: "Fish", Pages: 1, Total: 1,
		Almanaxes: []*amqp.Almanax{almanaxOf(day(2024, time.March, 3), "Bonus")},
	}

	embed := singleEmbed(t, MapAlmanaxEffectsToWebhook(answer, discordgo.EnglishGB, fakeEmojis{}).Embeds)

	assert.Equal(t, &discordgo.MessageEmbedAuthor{}, embed.Author)
}

func TestMapAlmanaxResourceToWebhook_MissingSourceLeavesAuthorEmpty(t *testing.T) {
	answer := &amqp.EncyclopediaAlmanaxResourceAnswer{Duration: 7}

	embed := singleEmbed(t, MapAlmanaxResourceToWebhook(answer, 1, discordgo.EnglishGB, fakeEmojis{}).Embeds)

	assert.Equal(t, &discordgo.MessageEmbedAuthor{}, embed.Author)
}

func assertOptions(t *testing.T, options []discordgo.SelectMenuOption, values []int64, selected string) {
	t.Helper()
	require.Len(t, options, len(values))
	for i, value := range values {
		expected := strconv.FormatInt(value, 10)
		assert.Equal(t, expected, options[i].Value)
		assert.Equal(t, expected == selected, options[i].Default, "option %s", expected)
		assert.Contains(t, options[i].Label, expected)
		assertRendered(t, options[i].Label)
	}
}
