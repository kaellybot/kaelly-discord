package guilds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	guildRepo "github.com/kaellybot/kaelly-discord/repositories/guilds"
)

type Service interface {
	GetServer(guildID, channelID string) (*entities.Server, error)
}

type Impl struct {
	guildRepo guildRepo.Repository
}
