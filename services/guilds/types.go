package guilds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	guildRepo "github.com/kaellybot/kaelly-discord/repositories/guilds"
)

type Service interface {
	Exists(guildID string) (bool, error)
	GetServer(guildID, channelID string) (entities.Server, bool, error)
}

type Impl struct {
	guildRepo guildRepo.Repository
}
