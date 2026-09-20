package mappers

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setAnswer(equipmentCount int, bonusNumbers ...int64) *amqp.EncyclopediaItemAnswer {
	set := &amqp.EncyclopediaItemAnswer_Set{Id: "12", Name: "Gobball Set", Icon: "set.png", Level: 20}
	for i := range equipmentCount {
		set.Equipments = append(set.Equipments, &amqp.EncyclopediaItemAnswer_Set_Equipment{
			Id: fmt.Sprint(100 + i), Name: fmt.Sprintf("Piece-%c", 'a'+i), Level: int64(10 + i),
			Type: amqp.EquipmentType_HAT,
		})
	}
	for _, number := range bonusNumbers {
		set.Bonuses = append(set.Bonuses, &amqp.EncyclopediaItemAnswer_Set_Bonus{
			ItemNumber: number,
			Effects:    []*amqp.EncyclopediaItemAnswer_Effect{effect("vit", fmt.Sprintf("%d Vitality", number*10))},
		})
	}
	return &amqp.EncyclopediaItemAnswer{Set: set, Source: source()}
}

func mapSet(answer *amqp.EncyclopediaItemAnswer, itemNumber int) *discordgo.WebhookEdit {
	return MapSetToWebhookEdit(answer, itemNumber, fakeCharacteristics{}, fakeEmojis{}, amqp.Language_EN)
}

func TestMapSetListRequest_CopiesQuery(t *testing.T) {
	request := MapSetListRequest("gobball", "author", discordgo.French)

	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, request.Type)
	assert.Equal(t, amqp.Language_FR, request.Language)
	assert.Equal(t, "author", request.UserID)
	require.NotNil(t, request.EncyclopediaListRequest)
	assert.Equal(t, "gobball", request.EncyclopediaListRequest.Query)
	assert.Equal(t, amqp.EncyclopediaListRequest_SET, request.EncyclopediaListRequest.Type)
}

func TestMapSetToWebhookEdit_MatchingBonusIsShownAndSelected(t *testing.T) {
	edit := mapSet(setAnswer(3, 2, 3), 2)

	embed := singleEmbed(t, edit.Embeds)
	assert.Equal(t, "Gobball Set", embed.Title)
	assert.Contains(t, embed.Description, "20")
	assertRendered(t, embed.Description)
	assert.Equal(t, "set.png", embed.Thumbnail.URL)
	assert.Equal(t, constants.Color, embed.Color)
	assertSourceAuthor(t, embed.Author)
	assertDefaultFooter(t, embed.Footer)

	require.Len(t, embed.Fields, 3)
	items, ghost, bonus := embed.Fields[0], embed.Fields[1], embed.Fields[2]
	assertTranslated(t, discordgo.EnglishGB, "set.items.title", items.Name)
	assertInOrder(t, items.Value, "Piece-a, lvl 10", "Piece-b, lvl 11", "Piece-c, lvl 12")
	assert.True(t, items.Inline)
	assert.Equal(t, constants.InvisibleCharacter, ghost.Name, "a ghost field separates items from bonuses")
	assert.False(t, ghost.Inline)
	assert.Contains(t, bonus.Name, "2")
	assertRendered(t, bonus.Name)
	assert.Contains(t, bonus.Value, "**20** Vitality")
	assert.True(t, bonus.Inline)

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 2)
	bonusMenu := selectMenu(t, rows[0])
	setID, ok := contract.ExtractSetBonusCustomID(bonusMenu.CustomID)
	require.True(t, ok, bonusMenu.CustomID)
	assert.Equal(t, "12", setID)
	assertTranslated(t, discordgo.EnglishGB, "set.effects.placeholder", bonusMenu.Placeholder)
	require.Len(t, bonusMenu.Options, 2)
	assert.Equal(t, "2", bonusMenu.Options[0].Value)
	assert.True(t, bonusMenu.Options[0].Default)
	assert.Equal(t, "bonus:2", bonusMenu.Options[0].Emoji.Name)
	assert.Contains(t, bonusMenu.Options[0].Label, "2 / 3", "options show the count against the max bonus")
	assert.Equal(t, "3", bonusMenu.Options[1].Value)
	assert.False(t, bonusMenu.Options[1].Default)

	itemMenu := selectMenu(t, rows[1])
	itemType, ok := contract.ExtractItemCustomID(itemMenu.CustomID)
	require.True(t, ok, itemMenu.CustomID)
	assert.Equal(t, amqp.ItemType_EQUIPMENT_TYPE.String(), itemType)
	assertTranslated(t, discordgo.EnglishGB, "set.items.placeholder", itemMenu.Placeholder)
	assert.Equal(t, discordgo.SelectMenuOption{Label: "Piece-a", Value: "100",
		Emoji: &discordgo.ComponentEmoji{Name: "equip:HAT"}}, itemMenu.Options[0])
	require.Len(t, itemMenu.Options, 3)
}

func TestMapSetToWebhookEdit_UnmatchedItemNumberFallsBackToHighestBonus(t *testing.T) {
	for _, itemNumber := range []int{1, 7} {
		t.Run(fmt.Sprint(itemNumber), func(t *testing.T) {
			edit := mapSet(setAnswer(3, 3, 5, 2), itemNumber)

			embed := singleEmbed(t, edit.Embeds)
			assert.Contains(t, embed.Fields[len(embed.Fields)-1].Value, "**50** Vitality")
			for _, option := range selectMenu(t, actionRows(t, edit.Components)[0]).Options {
				assert.Equal(t, option.Value == "5", option.Default, option.Value)
			}
		})
	}
}

func TestMapSetToDefaultWebhookEdit_UsesEquipmentCount(t *testing.T) {
	edit := MapSetToDefaultWebhookEdit(setAnswer(3, 2, 3, 4), fakeCharacteristics{}, fakeEmojis{}, amqp.Language_EN)

	for _, option := range selectMenu(t, actionRows(t, edit.Components)[0]).Options {
		assert.Equal(t, option.Value == "3", option.Default, option.Value)
	}
	embed := singleEmbed(t, edit.Embeds)
	assert.Contains(t, embed.Fields[len(embed.Fields)-1].Value, "**30** Vitality")
}

func TestMapSetToWebhookEdit_NoBonusesOnlyListsItems(t *testing.T) {
	edit := mapSet(setAnswer(constants.MaxEquipmentPerField+1), 0)

	embed := singleEmbed(t, edit.Embeds)
	require.Len(t, embed.Fields, 2, "items overflow onto a second field, no ghost nor bonus field")
	assertTranslated(t, discordgo.EnglishGB, "set.items.title", embed.Fields[0].Name)
	assert.Equal(t, constants.InvisibleCharacter, embed.Fields[1].Name)
	assert.Contains(t, embed.Fields[1].Value, "Piece-i")
	assert.True(t, embed.Fields[1].Inline)

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 1, "no bonus menu without bonuses")
	assert.Len(t, selectMenu(t, rows[0]).Options, constants.MaxEquipmentPerField+1)
}

func TestMapSetToWebhookEdit_BonusWithoutEffectsAddsNoField(t *testing.T) {
	answer := setAnswer(2, 2)
	answer.Set.Bonuses[0].Effects = nil

	edit := mapSet(answer, 2)

	embed := singleEmbed(t, edit.Embeds)
	require.Len(t, embed.Fields, 1)
	menu := selectMenu(t, actionRows(t, edit.Components)[0])
	assert.True(t, menu.Options[0].Default)
}

func TestMapSetToWebhookEdit_CosmeticSetLinksToCosmeticItems(t *testing.T) {
	answer := setAnswer(1)
	answer.Set.IsCosmetic = true

	rows := actionRows(t, mapSet(answer, 1).Components)

	itemType, ok := contract.ExtractItemCustomID(selectMenu(t, rows[len(rows)-1]).CustomID)
	require.True(t, ok)
	assert.Equal(t, amqp.ItemType_COSMETIC_TYPE.String(), itemType)
}
