package align

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	slashCommandName  = "align"
	userCommandName   = "Alignments"
	getSubCommandName = "get"
	setSubCommandName = "set"

	cityOptionName   = "city"
	orderOptionName  = "order"
	levelOptionName  = "level"
	serverOptionName = "server"

	alignRequestRoutingKey = "requests.books"

	memberListLimit   = 1000
	believerListLimit = 30

	userProperty = "user"
)

var (
	minLevel float64 = constants.AlignmentMinLevel
	maxLevel float64 = constants.AlignmentMaxLevel
)

type AlignCommand struct {
	bookService    books.BookService
	guildService   guilds.GuildService
	serverService  servers.ServerService
	requestManager requests.RequestManager
}
