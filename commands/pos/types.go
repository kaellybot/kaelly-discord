package pos

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	portalRequestRoutingKey = "requests.portals"
)

type Command struct {
	commands.AbstractCommand
	guildService   guilds.Service
	portalService  portals.Service
	serverService  servers.Service
	requestManager requests.RequestManager
	handlers       commands.DiscordHandlers
}
