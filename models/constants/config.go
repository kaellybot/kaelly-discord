package constants

import "github.com/rs/zerolog"

const (
	ConfigFileName = ".env"

	// Discord Bot Token
	Token = "TOKEN"

	// Shard ID. More on https://discord.com/developers/docs/topics/gateway#sharding
	ShardId = "SHARD_ID"

	// Total number of shards used to run the entire application
	ShardCount = "SHARD_COUNT"

	// MySQL URL with the following format: HOST:PORT
	MySqlUrl = "MYSQL_URL"

	// MySQL user
	MySqlUser = "MYSQL_USER"

	// MySQL password
	MySqlPassword = "MYSQL_PASSWORD"

	// MySQL database name
	MySqlDatabase = "MYSQL_DATABASE"

	// RabbitMQ address
	RabbitMqAddress = "RABBITMQ_ADDRESS"

	// Metric port
	MetricPort = "METRIC_PORT"

	// Zerolog values from [trace, debug, info, warn, error, fatal, panic]
	LogLevel = "LOG_LEVEL"

	// Boolean; used to register commands at development guild level or globally.
	Production = "PRODUCTION"
)

var (
	DefaultConfigValues = map[string]interface{}{
		Token:           "",
		ShardId:         0,
		ShardCount:      1,
		MySqlUrl:        "localhost:3306",
		MySqlUser:       "",
		MySqlPassword:   "",
		MySqlDatabase:   "kaellybot",
		RabbitMqAddress: "amqp://localhost:5672",
		MetricPort:      2112,
		LogLevel:        zerolog.InfoLevel.String(),
		Production:      false,
	}
)
