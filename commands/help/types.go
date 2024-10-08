package help

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
)

const (
	routingKey = "requests.help"

	menuCommandName = "menu"
)

type Command struct {
	commands.AbstractCommand
	broker   amqp.MessageBroker
	commands *[]commands.DiscordCommand
	handlers commands.DiscordHandlers
}
