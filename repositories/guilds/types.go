package guilds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	Exists(guildID string) (bool, error)
	GetServer(guildID, channelID string) (entities.Server, bool, error)
}

type Impl struct {
	db databases.MySQLConnection
}
