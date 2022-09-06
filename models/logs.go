package models

import "github.com/rs/zerolog"

const (
	LogCommand         = "command"
	LogGuildCount      = "guildCount"
	LogInteractionType = "interactionType"
	LogShard           = "shard"
	LogFileName        = "fileName"
	LogLocale          = "locale"
	LogPanic           = "panic"

	LogLevelFallback = zerolog.InfoLevel
)
