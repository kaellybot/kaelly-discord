package emojis

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/emojis"
)

type Service interface {
	GetMiscStringEmoji(emojiID constants.EmojiMiscID) string
	GetMiscEmoji(emojiID constants.EmojiMiscID) *discordgo.ComponentEmoji
	GetEquipmentStringEmoji(equipmentType amqp.EquipmentType) string
	GetEquipmentEmoji(equipmentType amqp.EquipmentType) *discordgo.ComponentEmoji
	GetItemTypeStringEmoji(itemType amqp.ItemType) string
	GetItemTypeEmoji(itemType amqp.ItemType) *discordgo.ComponentEmoji
	GetSetBonusEmoji(equipedItemNumber int) *discordgo.ComponentEmoji
}

type Impl struct {
	emojiStore map[constants.EmojiType]emojiStore
	repository repository.Repository
}

type emojiStore map[string]entities.Emoji
