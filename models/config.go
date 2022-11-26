package models

import "github.com/rs/zerolog"

const (
	ConfigFileName = ".env"

	// Discord Bot Token
	Token = "TOKEN"

	// Shard ID. More on https://discord.com/developers/docs/topics/gateway#sharding
	ShardId = "SHARD_ID"

	// Total number of shards used to run the entire application
	ShardCount = "SHARD_COUNT"

	// Zerolog values from [trace, debug, info, warn, error, fatal, panic]
	LogLevel = "LOG_LEVEL"

	// Boolean; used to register commands at development guild level or globally.
	Production = "PRODUCTION"
)

var (
	DefaultConfigValues = map[string]interface{}{
		Token:      "",
		ShardId:    0,
		ShardCount: 1,
		LogLevel:   zerolog.InfoLevel.String(),
		Production: false,
	}
)
