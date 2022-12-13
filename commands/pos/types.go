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

type PosCommand struct {
	guildService   guilds.GuildService
	portalService  portals.PortalService
	serverService  servers.ServerService
	requestManager requests.RequestManager
}
