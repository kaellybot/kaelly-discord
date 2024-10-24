package about

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/emojis"
)

const (
	routingKey = "requests.about"
)

type Command struct {
	commands.AbstractCommand
	broker       amqp.MessageBroker
	emojiService emojis.Service
}
