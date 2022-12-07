package constants

import (
	"github.com/spf13/viper"
)

func GetRabbitMQClientId() string {
	return Name + "-" + viper.GetString(ShardId)
}
