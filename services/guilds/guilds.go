package guilds

import "github.com/kaellybot/kaelly-discord/models/entities"

type GuildService interface {
	GetServer() *entities.Server
}

type GuildServiceImpl struct {
}

func New() *GuildServiceImpl {
	return &GuildServiceImpl{}
}

func (service *GuildServiceImpl) GetServer() *entities.Server {
	// TODO
	return nil
}
