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
	ErrCannotInstantiateApp = errors.New("Cannot instantiate application")
)

type ApplicationInterface interface {
	Run() error
	Shutdown()
}

type Application struct {
	db             databases.MySQLConnection
	broker         amqp.MessageBrokerInterface
	guildService   guilds.GuildService
	bookService    books.BookService
	portalService  portals.PortalService
	serverService  servers.ServerService
	discordService discord.DiscordService
	slashCommands  []commands.SlashCommand
	userCommands   []commands.UserCommand
	requestManager requests.RequestManager
}
