package mappers

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func areaEffect(id string, effectType constants.WeaponAreaType, order int) entities.WeaponAreaEffect {
	return entities.WeaponAreaEffect{ID: id, Type: effectType, Order: order,
		Labels: []entities.WeaponAreaEffectLabel{{WeaponAreaEffectID: id, Locale: amqp.Language_EN, Label: "L-" + id}}}
}

func itemEquipments() fakeEquipments {
	return fakeEquipments{effects: map[string]entities.WeaponAreaEffect{
		"los":     areaEffect("los", constants.Casting, 2),
		"line":    areaEffect("line", constants.Casting, 1),
		"mod":     areaEffect("mod", constants.Casting, 3),
		"circle":  areaEffect("circle", constants.Area, 2),
		"cross":   areaEffect("cross", constants.Area, 1),
		"strange": areaEffect("strange", constants.WeaponAreaType("other"), 0),
	}}
}

func equipmentCondition(name, operator string, value int64) *amqp.EncyclopediaItemAnswer_Conditions {
	return &amqp.EncyclopediaItemAnswer_Conditions{
		Condition: &amqp.EncyclopediaItemAnswer_Conditions_Condition{
			Operator: operator, Value: value,
			Element: &amqp.EncyclopediaItemAnswer_Conditions_Condition_Element{Name: name},
		},
	}
}

func ingredients(count int) []*amqp.EncyclopediaItemAnswer_Recipe_Ingredient {
	result := make([]*amqp.EncyclopediaItemAnswer_Recipe_Ingredient, 0, count)
	for i := range count {
		result = append(result, &amqp.EncyclopediaItemAnswer_Recipe_Ingredient{
			Id: fmt.Sprint(i), Name: fmt.Sprintf("Ingredient%d", i), Quantity: int64(i + 1),
			Type: amqp.ItemType_RESOURCE_TYPE,
		})
	}
	return result
}

func weaponAnswer() *amqp.EncyclopediaItemAnswer {
	return &amqp.EncyclopediaItemAnswer{
		Source: source(),
		Equipment: &amqp.EncyclopediaItemAnswer_Equipment{
			Id: "44", Name: "Crackler Hammer", Icon: "hammer.png", Level: 60,
			Type: &amqp.EncyclopediaItemAnswer_Equipment_Type{
				ItemType: amqp.ItemType_EQUIPMENT_TYPE, EquipmentType: amqp.EquipmentType_HAMMER,
				EquipmentLabel: "Hammer",
			},
			Set: &amqp.EncyclopediaItemAnswer_Equipment_SetFamily{Id: "9", Name: "Crackler Set"},
			Characteristics: &amqp.EncyclopediaItemAnswer_Equipment_Characteristics{
				Cost: 4, MinRange: 1, MaxRange: 2, MaxCastPerTurn: 1,
				CriticalRate: 5, CriticalBonus: 10, AreaEffectIds: []string{"circle"},
			},
			WeaponEffects: []*amqp.EncyclopediaItemAnswer_Effect{effect("earth", "10 to 20 (earth damage)")},
			Effects:       []*amqp.EncyclopediaItemAnswer_Effect{effect("vit", "100 Vitality")},
			Conditions:    equipmentCondition("Strength", ">", 50),
			Recipe: &amqp.EncyclopediaItemAnswer_Recipe{
				Ingredients: ingredients(2),
			},
		},
	}
}

func mapItem(answer *amqp.EncyclopediaItemAnswer, isRecipe bool) *discordgo.WebhookEdit {
	return MapItemToWebhookEdit(answer, isRecipe, fakeCharacteristics{}, itemEquipments(),
		fakeEmojis{}, amqp.Language_EN)
}

func TestMapItemRequests_CopyFields(t *testing.T) {
	list := MapItemListRequest("hammer", "author", discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_REQUEST, list.Type)
	assert.Equal(t, amqp.Language_FR, list.Language)
	assert.Equal(t, "author", list.UserID)
	require.NotNil(t, list.EncyclopediaListRequest)
	assert.Equal(t, "hammer", list.EncyclopediaListRequest.Query)
	assert.Equal(t, amqp.EncyclopediaListRequest_ITEM, list.EncyclopediaListRequest.Type)

	item := MapItemRequest("44", true, amqp.ItemType_COSMETIC_TYPE, "author", discordgo.German)
	assert.Equal(t, amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_REQUEST, item.Type)
	assert.Equal(t, amqp.Language_DE, item.Language)
	require.NotNil(t, item.EncyclopediaItemRequest)
	assert.Equal(t, "44", item.EncyclopediaItemRequest.Query)
	assert.True(t, item.EncyclopediaItemRequest.IsID)
	assert.Equal(t, amqp.ItemType_COSMETIC_TYPE, item.EncyclopediaItemRequest.Type)
}

func TestMapItemToWebhookEdit_EffectsViewShowsCharacteristicsEffectsAndConditions(t *testing.T) {
	edit := mapItem(weaponAnswer(), false)

	embed := singleEmbed(t, edit.Embeds)
	assert.Equal(t, "Crackler Hammer", embed.Title)
	assert.Contains(t, embed.Description, "60")
	assert.Contains(t, embed.Description, "<equip:HAMMER>")
	assert.Contains(t, embed.Description, "Hammer")
	assert.Equal(t, "hammer.png", embed.Thumbnail.URL)
	assert.Equal(t, constants.Color, embed.Color)
	assertSourceAuthor(t, embed.Author)
	assertDefaultFooter(t, embed.Footer)

	require.Len(t, embed.Fields, 4)
	charac, weaponEffects, effects, conditions := embed.Fields[0], embed.Fields[1], embed.Fields[2], embed.Fields[3]
	assertTranslated(t, discordgo.EnglishGB, "item.characteristics.title", charac.Name)
	assertRendered(t, charac.Value)
	for _, fragment := range []string{"4 <misc:cost>", "1 - 2 <misc:range>", "5% <misc:critical>", "+10",
		"L-circle <WeaponAreaEffect:circle>"} {
		assert.Contains(t, charac.Value, fragment)
	}
	assert.False(t, charac.Inline)
	assertTranslated(t, discordgo.EnglishGB, "item.weaponEffects.title", weaponEffects.Name)
	assert.Contains(t, weaponEffects.Value, "**10** to **20** (earth damage)")
	assert.False(t, weaponEffects.Inline)
	assertTranslated(t, discordgo.EnglishGB, "item.effects.title", effects.Name)
	assert.Contains(t, effects.Value, "**100** Vitality")
	assert.True(t, effects.Inline)
	assertTranslated(t, discordgo.EnglishGB, "item.conditions.title", conditions.Name)
	assert.Contains(t, conditions.Value, "Strength > 50")

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 1)
	btns := buttons(t, rows[0])
	require.Len(t, btns, 2)
	setID, ok := contract.ExtractSetCustomID(btns[0].CustomID)
	require.True(t, ok, btns[0].CustomID)
	assert.Equal(t, "9", setID)
	assert.Equal(t, "Crackler Set", btns[0].Label)
	assert.Equal(t, "item:SET_TYPE", btns[0].Emoji.Name)
	itemID, itemType, ok := contract.ExtractItemRecipeCustomID(btns[1].CustomID)
	require.True(t, ok, btns[1].CustomID)
	assert.Equal(t, "44", itemID)
	assert.Equal(t, amqp.ItemType_EQUIPMENT_TYPE.String(), itemType)
	assertTranslated(t, discordgo.EnglishGB, "item.recipe.button", btns[1].Label)
	assert.Equal(t, "misc:recipe", btns[1].Emoji.Name)
}

func TestMapItemToWebhookEdit_CriticalAndRangeAreOnlyDetailedWhenRelevant(t *testing.T) {
	answer := weaponAnswer()
	answer.Equipment.Characteristics.CriticalRate = 0
	answer.Equipment.Characteristics.MaxRange = 1
	answer.Equipment.Characteristics.AreaEffectIds = nil

	embed := singleEmbed(t, mapItem(answer, false).Embeds)

	charac := embed.Fields[0].Value
	assertRendered(t, charac)
	assert.NotContains(t, charac, "<misc:critical>")
	assert.NotContains(t, charac, "1 - 1")
	assert.Contains(t, charac, "1 <misc:range>")
	assert.NotContains(t, charac, "WeaponAreaEffect")
}

func TestMapItemToWebhookEdit_OnlyGenericEffectsGivesNoWeaponEffectField(t *testing.T) {
	answer := weaponAnswer()
	answer.Equipment.Characteristics = nil
	answer.Equipment.WeaponEffects = nil
	answer.Equipment.Conditions = nil
	answer.Equipment.Effects = make([]*amqp.EncyclopediaItemAnswer_Effect, 0)
	for i := range constants.MaxCharacterPerField + 1 {
		answer.Equipment.Effects = append(answer.Equipment.Effects, effect("x", fmt.Sprintf("Effect-%c", 'a'+i)))
	}

	embed := singleEmbed(t, mapItem(answer, false).Embeds)

	require.Len(t, embed.Fields, 2, "effects overflow onto a second field")
	assertTranslated(t, discordgo.EnglishGB, "item.effects.title", embed.Fields[0].Name)
	assert.Equal(t, constants.InvisibleCharacter, embed.Fields[1].Name)
	assert.Contains(t, embed.Fields[0].Value, "Effect-j")
	assert.NotContains(t, embed.Fields[0].Value, "Effect-k")
	assert.Contains(t, embed.Fields[1].Value, "Effect-k")
}

func TestMapItemToWebhookEdit_BareEquipmentHasNoFieldsNorComponents(t *testing.T) {
	answer := &amqp.EncyclopediaItemAnswer{Source: source(), Equipment: &amqp.EncyclopediaItemAnswer_Equipment{
		Id: "1", Name: "Ring", Level: 1,
		Type: &amqp.EncyclopediaItemAnswer_Equipment_Type{EquipmentLabel: "Ring"},
	}}

	for _, isRecipe := range []bool{false, true} {
		edit := mapItem(answer, isRecipe)
		embed := singleEmbed(t, edit.Embeds)
		assert.Empty(t, embed.Fields)
		assertRendered(t, embed.Description)
		assert.Nil(t, edit.Components)
	}
}

func TestMapItemToWebhookEdit_RecipeViewSlicesIngredientsAndOffersEffects(t *testing.T) {
	answer := weaponAnswer()
	answer.Equipment.Set = nil
	answer.Equipment.Recipe.Ingredients = ingredients(constants.MaxIngredientsPerField + 1)

	edit := mapItem(answer, true)

	embed := singleEmbed(t, edit.Embeds)
	require.Len(t, embed.Fields, 2)
	assertTranslated(t, discordgo.EnglishGB, "item.recipe.title", embed.Fields[0].Name)
	assert.Equal(t, constants.InvisibleCharacter, embed.Fields[1].Name)
	assert.True(t, embed.Fields[0].Inline)
	assertInOrder(t, embed.Fields[0].Value, "<item:RESOURCE_TYPE> **1**x Ingredient0",
		"<item:RESOURCE_TYPE> **8**x Ingredient7")
	assert.Contains(t, embed.Fields[1].Value, "**9**x Ingredient8")

	btns := buttons(t, actionRows(t, edit.Components)[0])
	require.Len(t, btns, 1)
	itemID, itemType, ok := contract.ExtractItemEffectsCustomID(btns[0].CustomID)
	require.True(t, ok, btns[0].CustomID)
	assert.Equal(t, "44", itemID)
	assert.Equal(t, amqp.ItemType_EQUIPMENT_TYPE.String(), itemType)
	assertTranslated(t, discordgo.EnglishGB, "item.effects.button", btns[0].Label)
	assert.Equal(t, "misc:effect", btns[0].Emoji.Name)
}

func TestMapItemToWebhookEdit_RecipeViewWithoutEffectsOffersRecipeAgain(t *testing.T) {
	answer := weaponAnswer()
	answer.Equipment.Set = nil
	answer.Equipment.WeaponEffects = nil
	answer.Equipment.Effects = nil

	edit := mapItem(answer, true)

	// Current behavior: with no effects to switch to, the recipe view still
	// offers a "recipe" button pointing to the view already displayed.
	btns := buttons(t, actionRows(t, edit.Components)[0])
	require.Len(t, btns, 1)
	_, _, ok := contract.ExtractItemRecipeCustomID(btns[0].CustomID)
	assert.True(t, ok, btns[0].CustomID)
}

func TestMapItemToWebhookEdit_RecipeViewWithoutRecipeHasNoFields(t *testing.T) {
	answer := weaponAnswer()
	answer.Equipment.Recipe = nil

	embed := singleEmbed(t, mapItem(answer, true).Embeds)

	assert.Empty(t, embed.Fields)
}

func TestGetConditions_FlattensAndAndJoinsOr(t *testing.T) {
	or := func(children ...*amqp.EncyclopediaItemAnswer_Conditions) *amqp.EncyclopediaItemAnswer_Conditions {
		return &amqp.EncyclopediaItemAnswer_Conditions{Relation: amqp.EncyclopediaItemAnswer_Conditions_OR,
			Children: children}
	}
	and := func(children ...*amqp.EncyclopediaItemAnswer_Conditions) *amqp.EncyclopediaItemAnswer_Conditions {
		return &amqp.EncyclopediaItemAnswer_Conditions{Relation: amqp.EncyclopediaItemAnswer_Conditions_AND,
			Children: children}
	}
	orWord := "or"

	cases := map[string]struct {
		conditions *amqp.EncyclopediaItemAnswer_Conditions
		expected   []string
	}{
		"empty": {conditions: &amqp.EncyclopediaItemAnswer_Conditions{}, expected: []string{}},
		"leaf":  {conditions: equipmentCondition("Level", "<", 100), expected: []string{"Level < 100"}},
		"and": {
			conditions: and(equipmentCondition("A", ">", 1), equipmentCondition("B", "<", 2)),
			expected:   []string{"A > 1", "B < 2"},
		},
		"or": {
			conditions: or(equipmentCondition("A", ">", 1), equipmentCondition("B", "<", 2)),
			expected:   []string{"A > 1 " + orWord + " B < 2"},
		},
		"and of or": {
			conditions: and(or(equipmentCondition("A", ">", 1), or(equipmentCondition("B", "<", 2),
				equipmentCondition("C", "=", 3))), equipmentCondition("D", ">", 4)),
			expected: []string{"A > 1 or B < 2 or C = 3", "D > 4"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getConditions(tc.conditions, discordgo.EnglishGB))
		})
	}

	french := getConditions(or(equipmentCondition("A", ">", 1), equipmentCondition("B", "<", 2)), discordgo.French)
	require.Len(t, french, 1)
	assertTranslated(t, discordgo.French, "item.conditions.relation.or", "ou")
	assert.Equal(t, "A > 1 ou B < 2", french[0])
}

func TestMapWeaponAreaEffect_PairsCastingsThenListsAreasInOrder(t *testing.T) {
	labels := mapWeaponAreaEffect([]string{"circle", "mod", "unknown", "los", "strange", "line", "cross"},
		itemEquipments(), fakeEmojis{}, discordgo.EnglishGB)

	require.Len(t, labels, 4)
	assertInOrder(t, labels[0], "*L-line*", "*L-los*")
	assert.Contains(t, labels[1], "*L-mod*", "odd casting is shown alone")
	assert.NotContains(t, labels[1], "L-los")
	assert.Contains(t, labels[2], "L-cross <WeaponAreaEffect:cross>")
	assert.Contains(t, labels[3], "L-circle <WeaponAreaEffect:circle>")
	for _, label := range labels {
		assertRendered(t, label)
		assert.NotContains(t, label, "strange")
	}
}

func TestMapWeaponAreaEffect_NoEffects(t *testing.T) {
	labels := mapWeaponAreaEffect(nil, itemEquipments(), fakeEmojis{}, discordgo.EnglishGB)
	assert.NotNil(t, labels)
	assert.Empty(t, labels)
}
