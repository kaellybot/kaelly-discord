package mappers

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

func MapGuildCreateNews(guildID, guildName string, memberCount int) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_NEWS_GUILD,
		Language: amqp.Language_ANY,
		Game:     constants.GetGame().AMQPGame,
		NewsGuildMessage: &amqp.NewsGuildMessage{
			Id:          guildID,
			Name:        guildName,
			MemberCount: int64(memberCount),
			Event:       amqp.NewsGuildMessage_CREATE,
		},
	}
}

func MapGuildDeleteNews(guildID, guildName string, memberCount int) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_NEWS_GUILD,
		Language: amqp.Language_ANY,
		Game:     constants.GetGame().AMQPGame,
		NewsGuildMessage: &amqp.NewsGuildMessage{
			Id:          guildID,
			Name:        guildName,
			MemberCount: int64(memberCount),
			Event:       amqp.NewsGuildMessage_DELETE,
		},
	}
}
