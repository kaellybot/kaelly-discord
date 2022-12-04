package models

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/spf13/viper"
)

const (
	answersRoutingKey = "answers.*"
	answersQueueName  = "answers"
)

func GetRabbitMQClientId() string {
	return Name + "-" + viper.GetString(ShardId)
}

func GetBindings() []amqp.Binding {
	return []amqp.Binding{
		{
			Exchange:   amqp.ExchangeAnswer,
			RoutingKey: answersRoutingKey,
			Queue:      answersQueueName,
		},
	}
}
