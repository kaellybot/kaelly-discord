package application

import (
	"errors"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/discord"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

var (
	ErrCannotInstantiateApp = errors.New("cannot instantiate application")
)

type Application interface {
	Run() error
	Shutdown()
}

type Impl struct {
	db             databases.MySQLConnection
	broker         amqp.MessageBroker
	guildService   guilds.Service
	bookService    books.Service
	portalService  portals.Service
	serverService  servers.Service
	discordService discord.Service
	slashCommands  []commands.SlashCommand
	userCommands   []commands.UserCommand
	requestManager requests.RequestManager
}
