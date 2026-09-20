package mappers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func alignBooks() fakeBooks {
	return fakeBooks{
		cities: []entities.City{
			{ID: "bonta", Icon: "bonta.png", Type: constants.CityTypeLight, Color: 111, Labels: []entities.CityLabel{
				{CityID: "bonta", Locale: amqp.Language_EN, Label: "Bonta"},
				{CityID: "bonta", Locale: amqp.Language_FR, Label: "Bonta-FR"},
			}},
			{ID: "brakmar", Icon: "brakmar.png", Type: constants.CityTypeDark, Color: 222, Labels: []entities.CityLabel{
				{CityID: "brakmar", Locale: amqp.Language_EN, Label: "Brakmar"},
			}},
		},
		orders: []entities.Order{
			{ID: "spirit", Labels: []entities.OrderLabel{{OrderID: "spirit", Locale: amqp.Language_EN, Label: "Spirit"}}},
			{ID: "heart", Labels: []entities.OrderLabel{{OrderID: "heart", Locale: amqp.Language_EN, Label: "Heart"}}},
		},
	}
}

func TestMapBookAlignRequests_CopyFieldsAndComputeOffset(t *testing.T) {
	book := MapBookAlignGetBookRequest("bonta", "spirit", "srv", 2, []string{"u1", "u2"}, "author", discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_ALIGN_GET_BOOK_REQUEST, book.Type)
	assert.Equal(t, amqp.Language_FR, book.Language)
	assert.Equal(t, "author", book.UserID)
	require.NotNil(t, book.AlignGetBookRequest)
	assert.Equal(t, []string{"u1", "u2"}, book.AlignGetBookRequest.UserIds)
	assert.Equal(t, "bonta", book.AlignGetBookRequest.CityId)
	assert.Equal(t, "spirit", book.AlignGetBookRequest.OrderId)
	assert.Equal(t, "srv", book.AlignGetBookRequest.ServerId)
	assert.Equal(t, int64(2*constants.MaxBookRowPerEmbed), book.AlignGetBookRequest.Offset)
	assert.Equal(t, int64(constants.MaxBookRowPerEmbed), book.AlignGetBookRequest.Size)

	firstPage := MapBookAlignGetBookRequest("", "", "srv", 0, nil, "author", discordgo.French)
	assert.Zero(t, firstPage.AlignGetBookRequest.Offset)

	user := MapBookAlignGetUserRequest("target", "srv", "author", discordgo.EnglishUS)
	assert.Equal(t, amqp.RabbitMQMessage_ALIGN_GET_USER_REQUEST, user.Type)
	assert.Equal(t, amqp.Language_EN, user.Language)
	assert.Equal(t, "author", user.UserID)
	require.NotNil(t, user.AlignGetUserRequest)
	assert.Equal(t, "target", user.AlignGetUserRequest.UserId)
	assert.Equal(t, "srv", user.AlignGetUserRequest.ServerId)
}

func TestMapBookAlignSetRequest_UserIsAlsoTheAuthor(t *testing.T) {
	request := MapBookAlignSetRequest("user", "bonta", "spirit", "srv", 150, discordgo.German)

	assert.Equal(t, amqp.RabbitMQMessage_ALIGN_SET_REQUEST, request.Type)
	assert.Equal(t, amqp.Language_DE, request.Language)
	assert.Equal(t, "user", request.UserID)
	require.NotNil(t, request.AlignSetRequest)
	assert.Equal(t, "user", request.AlignSetRequest.UserId)
	assert.Equal(t, "bonta", request.AlignSetRequest.CityId)
	assert.Equal(t, "spirit", request.AlignSetRequest.OrderId)
	assert.Equal(t, "srv", request.AlignSetRequest.ServerId)
	assert.Equal(t, int64(150), request.AlignSetRequest.Level)
}

func TestBuildEmojiOrderID_PrefixesWithCityType(t *testing.T) {
	cases := map[string]struct {
		city     *entities.City
		expected string
	}{
		"no city":      {city: nil, expected: "spirit"},
		"light city":   {city: &entities.City{Type: constants.CityTypeLight}, expected: "lightspirit"},
		"dark city":    {city: &entities.City{Type: constants.CityTypeDark}, expected: "darkspirit"},
		"neutral city": {city: &entities.City{}, expected: "spirit"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, buildEmojiOrderID(tc.city, "spirit"))
		})
	}
}

func TestMapWinningCity_HighestTotalWinsTiesAreNeutral(t *testing.T) {
	service := alignBooks()
	cases := map[string]struct {
		values   map[string]int64
		expected string
	}{
		"nobody":                    {values: map[string]int64{}, expected: ""},
		"single city":               {values: map[string]int64{"bonta": 10}, expected: "bonta"},
		"highest wins":              {values: map[string]int64{"bonta": 10, "brakmar": 30}, expected: "brakmar"},
		"tie is neutral":            {values: map[string]int64{"bonta": 20, "brakmar": 20}, expected: ""},
		"unknown winner is neutral": {values: map[string]int64{"bonta": 10, "ghost": 99}, expected: ""},
		"tie below winner ignored": {
			values:   map[string]int64{"bonta": 50, "brakmar": 20, "ghost": 20},
			expected: "bonta",
		},
		"only zero levels": {values: map[string]int64{"bonta": 0}, expected: ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Map iteration order is random: repeat to shake out order dependence.
			for range 50 {
				city := mapWinningCity(tc.values, service)
				assert.Equal(t, tc.expected, city.ID)
				if tc.expected == "" {
					assert.Equal(t, entities.GetNeutralCity(), city)
				}
			}
		})
	}
}

func TestMapAlignBookToWebhook_ListsBelieversSortedAndColorsByWinningCity(t *testing.T) {
	service := alignBooks()
	serverService := fakeServers{servers: []entities.Server{testServer()}}
	answer := &amqp.AlignGetBookAnswer{ServerId: "srv", Page: 1, Pages: 3, Total: 40}
	believers := []constants.AlignmentUserLevel{
		{CityID: "brakmar", OrderID: "heart", Username: "Low", Level: 10},
		{CityID: "bonta", OrderID: "spirit", Username: "High", Level: 200},
		{CityID: "bonta", OrderID: "heart", Username: "TieBontaHeart", Level: 50},
		{CityID: "brakmar", OrderID: "spirit", Username: "TieBrakmar", Level: 50},
		{CityID: "bonta", OrderID: "spirit", Username: "TieBontaSpirit", Level: 50},
	}

	edit := MapAlignBookToWebhook(answer, believers, service, serverService, fakeEmojis{}, discordgo.EnglishGB)

	embed := singleEmbed(t, edit.Embeds)
	assertTranslated(t, discordgo.EnglishGB, "align.embed.believers.title", embed.Title)
	assertRendered(t, embed.Description)
	assert.Contains(t, embed.Description, "40")
	assert.Contains(t, embed.Description, "page 2 on 3", "page is displayed 1-based")
	// Level desc, then city emoji, then order emoji; light cities get prefixed order emojis.
	assertInOrder(t, embed.Description,
		"<City:bonta> <Order:lightspirit> lvl. 200: **High**",
		"<City:bonta> <Order:lightheart> lvl. 50: **TieBontaHeart**",
		"<City:bonta> <Order:lightspirit> lvl. 50: **TieBontaSpirit**",
		"<City:brakmar> <Order:darkspirit> lvl. 50: **TieBrakmar**",
		"<City:brakmar> <Order:darkheart> lvl. 10: **Low**",
	)
	assert.Equal(t, 111, embed.Color, "bonta totals 300 against 60")
	assert.Equal(t, "bonta.png", embed.Thumbnail.URL)
	assert.Equal(t, "Imagiro", embed.Footer.Text)
	assert.Equal(t, "srv.png", embed.Footer.IconURL)
}

func TestMapAlignBookToWebhook_NoBelieversIsNeutralAndUnknownServerUsesID(t *testing.T) {
	answer := &amqp.AlignGetBookAnswer{ServerId: "ghost", Pages: 0}

	edit := MapAlignBookToWebhook(answer, nil, alignBooks(), fakeServers{}, fakeEmojis{}, discordgo.EnglishGB)

	embed := singleEmbed(t, edit.Embeds)
	assertRendered(t, embed.Description)
	assert.NotContains(t, embed.Description, "**")
	assert.Equal(t, entities.GetNeutralCity().Color, embed.Color)
	assert.Equal(t, entities.GetNeutralCity().Icon, embed.Thumbnail.URL)
	assert.Equal(t, "ghost", embed.Footer.Text)
	assert.Empty(t, embed.Footer.IconURL)

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 2, "a single page has no pagination row")
}

func TestMapAlignBookToWebhook_UnknownCityAndOrderStillListed(t *testing.T) {
	believers := []constants.AlignmentUserLevel{{CityID: "ghost", OrderID: "void", Username: "Lost", Level: 5}}

	edit := MapAlignBookToWebhook(&amqp.AlignGetBookAnswer{ServerId: "srv"}, believers,
		alignBooks(), fakeServers{}, fakeEmojis{}, discordgo.EnglishGB)

	embed := singleEmbed(t, edit.Embeds)
	assert.Contains(t, embed.Description, "<City:> <Order:> lvl. 5: **Lost**")
	assert.Equal(t, entities.GetNeutralCity().Color, embed.Color)
}

func TestMapAlignBookToComponents_PaginationAndFilterSelects(t *testing.T) {
	answer := &amqp.AlignGetBookAnswer{CityId: "bonta", OrderId: "heart", ServerId: "srv", Page: 1, Pages: 3}

	rows := actionRows(t, mapAlignBookToComponents(answer, discordgo.French, alignBooks(), fakeEmojis{}))

	require.Len(t, rows, 3)
	pagination := buttons(t, rows[0])
	require.Len(t, pagination, 2)
	for i, expectedPage := range []int{0, 2} {
		cityID, orderID, serverID, page, ok := contract.ExtractAlignBookPageCustomID(pagination[i].CustomID)
		require.True(t, ok, pagination[i].CustomID)
		assert.Equal(t, []any{"bonta", "heart", "srv", expectedPage}, []any{cityID, orderID, serverID, page})
	}

	cityMenu := selectMenu(t, rows[1])
	orderID, serverID, ok := contract.ExtractAlignBookCityCustomID(cityMenu.CustomID)
	require.True(t, ok, cityMenu.CustomID)
	assert.Equal(t, "heart", orderID, "city menu keeps the order filter")
	assert.Equal(t, "srv", serverID)
	assertTranslated(t, discordgo.French, "align.embed.believers.placeholders.city", cityMenu.Placeholder)
	require.Len(t, cityMenu.Options, 3)
	all := cityMenu.Options[0]
	assert.Equal(t, contract.AlignAllValues, all.Value)
	assert.False(t, all.Default)
	assert.Equal(t, "misc:global", all.Emoji.Name)
	assertTranslated(t, discordgo.French, "align.embed.believers.placeholders.cities", all.Label)
	assert.Equal(t, discordgo.SelectMenuOption{Label: "Bonta-FR", Value: "bonta", Default: true,
		Emoji: &discordgo.ComponentEmoji{Name: "City:bonta"}}, cityMenu.Options[1])
	assert.Equal(t, discordgo.SelectMenuOption{Label: "Brakmar", Value: "brakmar", Default: false,
		Emoji: &discordgo.ComponentEmoji{Name: "City:brakmar"}}, cityMenu.Options[2],
		"missing French label falls back to English")

	orderMenu := selectMenu(t, rows[2])
	cityID, serverID, ok := contract.ExtractAlignBookOrderCustomID(orderMenu.CustomID)
	require.True(t, ok, orderMenu.CustomID)
	assert.Equal(t, "bonta", cityID, "order menu keeps the city filter")
	assert.Equal(t, "srv", serverID)
	require.Len(t, orderMenu.Options, 3)
	assert.Equal(t, contract.AlignAllValues, orderMenu.Options[0].Value)
	assert.False(t, orderMenu.Options[0].Default)
	assert.Equal(t, "spirit", orderMenu.Options[1].Value)
	assert.False(t, orderMenu.Options[1].Default)
	assert.Equal(t, "Order:lightspirit", orderMenu.Options[1].Emoji.Name, "order emojis follow the chosen city")
	assert.Equal(t, "heart", orderMenu.Options[2].Value)
	assert.True(t, orderMenu.Options[2].Default)
	assert.Equal(t, "Order:lightheart", orderMenu.Options[2].Emoji.Name)
}

func TestMapAlignBookToComponents_NoFilterSelectsAllAndUsesPlainOrderEmojis(t *testing.T) {
	answer := &amqp.AlignGetBookAnswer{ServerId: "srv", Pages: 1}

	rows := actionRows(t, mapAlignBookToComponents(answer, discordgo.EnglishGB, alignBooks(), fakeEmojis{}))

	require.Len(t, rows, 2)
	cityMenu, orderMenu := selectMenu(t, rows[0]), selectMenu(t, rows[1])
	assert.True(t, cityMenu.Options[0].Default)
	assert.True(t, orderMenu.Options[0].Default)
	for _, option := range append(cityMenu.Options[1:], orderMenu.Options[1:]...) {
		assert.False(t, option.Default, option.Value)
	}
	assert.Equal(t, "Order:spirit", orderMenu.Options[1].Emoji.Name)
	orderID, _, ok := contract.ExtractAlignBookCityCustomID(cityMenu.CustomID)
	require.True(t, ok)
	assert.Empty(t, orderID)
}

func TestMapAlignUserToEmbed_UsesNicknameAndSortsBeliefs(t *testing.T) {
	member := &discordgo.Member{GuildID: "g", Nick: "Nick", Avatar: "memberavatar",
		User: &discordgo.User{ID: "u", Username: "user", Avatar: "useravatar"}}
	experiences := []*amqp.AlignGetUserAnswer_AlignExperience{
		{CityId: "brakmar", OrderId: "heart", Level: 100},
		{CityId: "bonta", OrderId: "spirit", Level: 100},
		{CityId: "bonta", OrderId: "heart", Level: 20},
	}

	embeds := MapAlignUserToEmbed(experiences, member, "srv", alignBooks(), fakeEmojis{},
		fakeServers{servers: []entities.Server{testServer()}}, amqp.Language_FR)

	embed := singleEmbed(t, embeds)
	assert.Contains(t, embed.Title, "Nick")
	assertRendered(t, embed.Title)
	assertInOrder(t, embed.Description,
		"<City:bonta> <Order:lightspirit>",
		"<City:brakmar> <Order:darkheart>",
		"<City:bonta> <Order:lightheart>",
	)
	assert.Equal(t, member.AvatarURL(""), embed.Thumbnail.URL)
	assert.Contains(t, embed.Thumbnail.URL, "memberavatar")
	assert.Equal(t, 111, embed.Color, "bonta totals 120 against 100")
	assert.Equal(t, "Imagiro-FR", embed.Footer.Text)
}

func TestMapAlignUserToEmbed_MemberWithoutUserDegradesGracefully(t *testing.T) {
	member := &discordgo.Member{Nick: "Nick"}

	embeds := MapAlignUserToEmbed(nil, member, "srv", alignBooks(), fakeEmojis{},
		fakeServers{servers: []entities.Server{testServer()}}, amqp.Language_EN)

	embed := singleEmbed(t, embeds)
	assert.Contains(t, embed.Title, "Nick")
	assert.Empty(t, embed.Thumbnail.URL)
}

func TestMapAlignUserToEmbed_FallsBackToUsernameUserAvatarAndEmptyEntities(t *testing.T) {
	member := &discordgo.Member{User: &discordgo.User{ID: "u", Username: "user", Avatar: "useravatar"}}

	experiences := []*amqp.AlignGetUserAnswer_AlignExperience{{CityId: "ghost", OrderId: "void", Level: 1}}

	embeds := MapAlignUserToEmbed(experiences, member, "ghost", alignBooks(), fakeEmojis{},
		fakeServers{}, amqp.Language_EN)

	embed := singleEmbed(t, embeds)
	assert.Contains(t, embed.Title, "user")
	assert.Contains(t, embed.Description, "<City:> <Order:>", "unknown city and order are still listed")
	assert.Contains(t, embed.Thumbnail.URL, "useravatar")
	assertRendered(t, embed.Description)
	assert.Equal(t, entities.GetNeutralCity().Color, embed.Color)
	assert.Equal(t, "ghost", embed.Footer.Text)
}
