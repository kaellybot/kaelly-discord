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
	GetEquipmentEmoji(equipmentType amqp.EquipmentType) *discordgo.ComponentEmoji
	GetSetBonusEmoji(equipedItemNumber, itemNumber int) *discordgo.ComponentEmoji
}

type Impl struct {
	emojiStore map[constants.EmojiType]emojiStore
	repository repository.Repository
}

type emojiStore map[string]entities.Emoji
