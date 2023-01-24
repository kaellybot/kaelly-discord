package job

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	slashCommandName  = "job"
	userCommandName   = "Jobs"
	getSubCommandName = "get"
	setSubCommandName = "set"

	jobOptionName    = "job"
	levelOptionName  = "level"
	serverOptionName = "server"

	jobRequestRoutingKey = "requests.books"

	memberListLimit    = 1000
	craftsmenListLimit = 30

	userProperty = "user"
)

var (
	minLevel float64 = constants.JobMinLevel
	maxLevel float64 = constants.JobMaxLevel
)

type JobCommand struct {
	bookService    books.BookService
	guildService   guilds.GuildService
	serverService  servers.ServerService
	requestManager requests.RequestManager
}
