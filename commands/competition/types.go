package competition

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	competitionRequestRoutingKey = "requests.competitions"
)

type Command struct {
	commands.AbstractCommand
	requestManager requests.RequestManager
	emojiService   emojis.Service
	handlers       commands.DiscordHandlers
}
