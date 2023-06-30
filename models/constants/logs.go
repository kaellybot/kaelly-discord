package constants

import "github.com/rs/zerolog"

const (
	LogCommand         = "command"
	LogCommandOption   = "option"
	LogGuildCount      = "guildCount"
	LogGuildID         = "guildID"
	LogChannelID       = "channelID"
	LogEntity          = "entity"
	LogAnkamaID        = "ankamaID"
	LogItemNumber      = "itemNumber"
	LogEmojiType       = "emojiType"
	LogInteractionType = "interactionType"
	LogShard           = "shard"
	LogFileName        = "fileName"
	LogLocale          = "locale"
	LogPanic           = "panic"

	LogLevelFallback = zerolog.InfoLevel
)
