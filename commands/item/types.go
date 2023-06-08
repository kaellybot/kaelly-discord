package item

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	configurationRequestRoutingKey = "requests.items"

	defaultIconSize = "128"
)

type Command struct {
	commands.AbstractCommand
	requestManager requests.RequestManager
	handlers       commands.DiscordHandlers
}
