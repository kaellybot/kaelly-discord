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
