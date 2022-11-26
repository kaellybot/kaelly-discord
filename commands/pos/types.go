package pos

import (
	"github.com/kaellybot/kaelly-discord/services/dimensions"
	"github.com/kaellybot/kaelly-discord/services/servers"
)

const (
	commandName         = "pos"
	dimensionOptionName = "dimension"
	serverOptionName    = "server"
)

type PosCommand struct {
	dimensionService dimensions.DimensionService
	serverService    servers.ServerService
}
