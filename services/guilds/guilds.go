package guilds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	guildRepo "github.com/kaellybot/kaelly-discord/repositories/guilds"
)

type GuildService interface {
	GetServer(guildId, channelId string) (*entities.Server, error)
}

type GuildServiceImpl struct {
	guildRepo guildRepo.GuildRepository
}

func New(guildRepo guildRepo.GuildRepository) *GuildServiceImpl {
	return &GuildServiceImpl{
		guildRepo: guildRepo,
	}
}

func (service *GuildServiceImpl) GetServer(guildId, channelId string) (*entities.Server, error) {
	server, err := service.guildRepo.GetServer(guildId, channelId)
	if err != nil {
		return nil, err
	}

	return server, nil
}
