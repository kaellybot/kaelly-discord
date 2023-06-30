package emojis

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/emojis"
)

type Service interface {
	GetEquipmentEmoji(equipmentType amqp.EquipmentType) entities.Emoji
	GetSetBonusEmoji(equipedItemNumber, itemNumber int) entities.Emoji
}

type Impl struct {
	emojiStore map[constants.EmojiType]emojiStore
	repository repository.Repository
}

type emojiStore map[string]entities.Emoji

const ( // TODO store in DB and implemnt service
	EmojiEffectID = "🔥"
	EmojiItemID   = "1124008991262003271"
	EmojiSetID    = "1123998108930551909"
	EmojiHatID    = "1124106262427222067"
	EmojiCloakID  = "1124106293263741058"
	EmojiBeltID   = "1124106327275348009"
	EmojiBootID   = "1124106363275067453"
	EmojiAmuletID = "1124106387841101964"
	EmojiRingID   = "1124106401120264332"
	EmojiWeaponID = "1124110029847535747"
	EmojiPetID    = "1124261009008365648"
)
