package job

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	jobRequestRoutingKey = "requests.books"

	memberListLimit    = 1000
	craftsmenListLimit = 30

	userProperty = "user"
)

type Command struct {
	commands.AbstractCommand
	bookService    books.Service
	guildService   guilds.Service
	serverService  servers.Service
	requestManager requests.RequestManager
	slashHandlers  commands.DiscordHandlers
	userHandlers   commands.DiscordHandlers
}
