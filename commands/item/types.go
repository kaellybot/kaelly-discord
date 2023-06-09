package item

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	itemRequestRoutingKey = "requests.encyclopedias"

	defaultIconSize = "128"
)

type Command struct {
	commands.AbstractCommand
	requestManager requests.RequestManager
	handlers       commands.DiscordHandlers
}
