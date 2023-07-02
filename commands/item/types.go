package item

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	itemRequestRoutingKey = "requests.encyclopedias"
	isRecipeProperty      = "isRecipe"
)

type Command struct {
	commands.AbstractCommand
	characService  characteristics.Service
	emojiService   emojis.Service
	requestManager requests.RequestManager
	handlers       commands.DiscordHandlers
}
