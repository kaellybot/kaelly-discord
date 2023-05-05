package constants

import (
	"github.com/spf13/viper"
)

func GetRabbitMQClientID() string {
	return Name + "-" + viper.GetString(ShardID)
}
