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

	// About command ID.
	AboutID = "ABOUT_ID"

	// Align command ID.
	AlignID = "ALIGN_ID"

	// Almanax command ID.
	AlmanaxID = "ALMANAX_ID"

	// Config command ID.
	ConfigID = "CONFIG_ID"
	// Help command ID.
	HelpID = "HELP_ID"

	// Item command ID.
	ItemID = "ITEM_ID"

	// Job command ID.
	JobID = "JOB_ID"

	// Map command ID.
	MapID = "MAP_ID"

	// Pos command ID.
	PosID = "POS_ID"

	// Set command ID.
	SetID = "SET_ID"

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
	defaultAboutID         = "1014249375154835557"
	defaultAlignID         = "1069057760269963295"
	defaultAlmanaxID       = "1177674483876761610"
	defaultConfigID        = "1055459522812067840"
	defaultHelpID          = "1190612462194663555"
	defaultItemID          = "1116290248587100251"
	defaultJobID           = "1062090620656681092"
	defaultMapID           = "1291722831767404667"
	defaultPosID           = "1020995396648054804"
	defaultSetID           = "1117887213481496607"
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
		AboutID:         defaultAboutID,
		AlignID:         defaultAlignID,
		AlmanaxID:       defaultAlmanaxID,
		ConfigID:        defaultConfigID,
		HelpID:          defaultHelpID,
		ItemID:          defaultItemID,
		JobID:           defaultJobID,
		MapID:           defaultMapID,
		PosID:           defaultPosID,
		SetID:           defaultSetID,
		ProbePort:       defaultProbePort,
		MetricPort:      defaultMetricPort,
		LogLevel:        defaultLogLevel.String(),
		Production:      defaultProduction,
	}
}
