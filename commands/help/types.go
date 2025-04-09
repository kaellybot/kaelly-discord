package help

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/emojis"
)

const (
	menuCommandName = "menu"
)

type Command struct {
	commands.AbstractCommand
	broker       amqp.MessageBroker
	commands     *[]commands.DiscordCommand
	handlers     commands.DiscordHandlers
	emojiService emojis.Service
}
