package constants

import "github.com/rs/zerolog"

const (
	LogCommand         = "command"
	LogCommandOption   = "option"
	LogGuildCount      = "guildCount"
	LogEntity          = "entity"
	LogInteractionType = "interactionType"
	LogShard           = "shard"
	LogFileName        = "fileName"
	LogLocale          = "locale"
	LogPanic           = "panic"

	LogLevelFallback = zerolog.InfoLevel
)
