package almanax

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	almanaxRequestRoutingKey = "requests.encyclopedias"
)

type Command struct {
	commands.AbstractCommand
	emojiService   emojis.Service
	requestManager requests.RequestManager
	handlers       commands.DiscordHandlers
}
