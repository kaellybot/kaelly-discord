package constants

import "github.com/rs/zerolog"

const (
	ConfigFileName = ".env"

	// Discord Bot Token.
	Token = "TOKEN"

	// Shard ID. More on https://discord.com/developers/docs/topics/gateway#sharding.
	ShardID = "SHARD_ID"

	// Total number of shards used to run the entire application.
	ShardCount = "SHARD_COUNT"

	// MySQL URL with the following format: HOST:PORT.
	MySQLURL = "MYSQL_URL"

	// MySQL user.
	MySQLUser = "MYSQL_USER"

	// MySQL password.
	MySQLPassword = "MYSQL_PASSWORD"

	// MySQL database name.
	MySQLDatabase = "MYSQL_DATABASE"

	// RabbitMQ address.
	RabbitMQAddress = "RABBITMQ_ADDRESS"

	// Probe port.
	ProbePort = "PROBE_PORT"

	// Metric port.
	MetricPort = "METRIC_PORT"

	// Zerolog values from [trace, debug, info, warn, error, fatal, panic].
	LogLevel = "LOG_LEVEL"

	// Boolean; used to register commands at development guild level or globally.
	Production = "PRODUCTION"

	// Default values.
	defaultToken           = ""
	defaultShardID         = 0
	defaultShardCount      = 1
	defaultMySQLURL        = "localhost:3306"
	defaultMySQLUser       = ""
	defaultMySQLPassword   = ""
	defaultMySQLDatabase   = "kaellybot"
	defaultRabbitMQAddress = "amqp://localhost:5672"
	defaultProbePort       = 9090
	defaultMetricPort      = 2112
	defaultLogLevel        = zerolog.InfoLevel
	defaultProduction      = false
)

func GetDefaultConfigValues() map[string]any {
	return map[string]any{
		Token:           defaultToken,
		ShardID:         defaultShardID,
		ShardCount:      defaultShardCount,
		MySQLURL:        defaultMySQLURL,
		MySQLUser:       defaultMySQLUser,
		MySQLPassword:   defaultMySQLPassword,
		MySQLDatabase:   defaultMySQLDatabase,
		RabbitMQAddress: defaultRabbitMQAddress,
		ProbePort:       defaultProbePort,
		MetricPort:      defaultMetricPort,
		LogLevel:        defaultLogLevel.String(),
		Production:      defaultProduction,
	}
}
