package constants

import "github.com/rs/zerolog"

const (
	LogCommand         = "command"
	LogCommandOption   = "option"
	LogGuildCount      = "guildCount"
	LogGuildID         = "guildID"
	LogChannelID       = "channelID"
	LogCustomID        = "customID"
	LogEntity          = "entity"
	LogAnkamaID        = "ankamaID"
	LogItemNumber      = "itemNumber"
	LogEmojiType       = "emojiType"
	LogRequestProperty = "requestProperty"
	LogRequestValue    = "requestValue"
	LogInteractionType = "interactionType"
	LogShard           = "shard"
	LogFileName        = "fileName"
	LogLocale          = "locale"
	LogPanic           = "panic"

	LogLevelFallback = zerolog.InfoLevel
)
