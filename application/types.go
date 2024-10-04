package application

import (
	"errors"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/services/discord"
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
	discordService discord.Service
	requestManager requests.RequestManager
}
