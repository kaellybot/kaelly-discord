package constants

import "github.com/rs/zerolog"

const (
	LogCommand              = "command"
	LogCommandOption        = "option"
	LogGuildCount           = "guildCount"
	LogGuildID              = "guildID"
	LogChannelID            = "channelID"
	LogCustomID             = "customID"
	LogEntity               = "entity"
	LogEntityCount          = "entityCount"
	LogAnkamaID             = "ankamaID"
	LogItemNumber           = "itemNumber"
	LogItemType             = "itemType"
	LogEmojiType            = "emojiType"
	LogWeaponAreaEffectType = "weaponAreaEffectType"
	LogRequestProperty      = "requestProperty"
	LogRequestValue         = "requestValue"
	LogInteractionType      = "interactionType"
	LogShard                = "shard"
	LogFileName             = "fileName"
	LogLocale               = "locale"
	LogPanic                = "panic"

	LogLevelFallback = zerolog.InfoLevel
)
