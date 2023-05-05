package pos

import (
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	commandName         = "pos"
	dimensionOptionName = "dimension"
	serverOptionName    = "server"

	portalRequestRoutingKey = "requests.portals"
)

type Command struct {
	guildService   guilds.Service
	portalService  portals.Service
	serverService  servers.Service
	requestManager requests.RequestManager
}
