package guilds

import "github.com/kaellybot/kaelly-discord/models"

type GuildService interface {
	GetServer() *models.Server
}

type GuildServiceImpl struct {
}

func New() *GuildServiceImpl {
	return &GuildServiceImpl{}
}

func (service *GuildServiceImpl) GetServer() *models.Server {
	// TODO
	return nil
}
