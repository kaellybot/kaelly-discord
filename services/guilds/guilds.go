package guilds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	guildRepo "github.com/kaellybot/kaelly-discord/repositories/guilds"
)

func New(guildRepo guildRepo.Repository) *Impl {
	return &Impl{
		guildRepo: guildRepo,
	}
}

func (service *Impl) GetServer(guildID, channelID string) (entities.Server, bool, error) {
	return service.guildRepo.GetServer(guildID, channelID)
}
