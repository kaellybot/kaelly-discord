package servers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/servers"
)

type ServerService interface {
	GetServers() []entities.Server
	FindServers(name string, locale discordgo.Locale) []entities.Server
}

type ServerServiceImpl struct {
	servers    []entities.Server
	repository repository.ServerRepository
}

func New(repository repository.ServerRepository) (*ServerServiceImpl, error) {
	servers, err := repository.GetServers()
	if err != nil {
		return nil, err
	}
	return &ServerServiceImpl{
		servers:    servers,
		repository: repository,
	}, nil
}

func (service *ServerServiceImpl) GetServers() []entities.Server {
	return service.servers
}

func (service *ServerServiceImpl) FindServers(name string, locale discordgo.Locale) []entities.Server {
	serversFound := make([]entities.Server, 0)
	cleanedName := strings.ToLower(name)

	// TODO not based on id

	for _, server := range service.servers {
		if strings.HasPrefix(strings.ToLower(server.Id), cleanedName) {
			serversFound = append(serversFound, server)
		}
	}

	return serversFound
}
