package servers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/servers"
	"github.com/kaellybot/kaelly-discord/utils/i18n"
)

type ServerService interface {
	GetServer(id string) (entities.Server, bool)
	GetServers() []entities.Server
	FindServers(name string, locale discordgo.Locale) []entities.Server
}

type ServerServiceImpl struct {
	serversMap map[string]entities.Server
	servers    []entities.Server
	repository repository.ServerRepository
}

func New(repository repository.ServerRepository) (*ServerServiceImpl, error) {
	servers, err := repository.GetServers()
	if err != nil {
		return nil, err
	}

	serversMap := make(map[string]entities.Server)
	for _, server := range servers {
		serversMap[server.Id] = server
	}

	return &ServerServiceImpl{
		serversMap: serversMap,
		servers:    servers,
		repository: repository,
	}, nil
}

func (service *ServerServiceImpl) GetServers() []entities.Server {
	return service.servers
}

func (service *ServerServiceImpl) GetServer(id string) (entities.Server, bool) {
	server, found := service.serversMap[id]
	return server, found
}

func (service *ServerServiceImpl) FindServers(name string, locale discordgo.Locale) []entities.Server {
	serversFound := make([]entities.Server, 0)
	cleanedName := strings.ToLower(name)

	// TODO normalize names

	for _, server := range service.servers {
		if strings.HasPrefix(strings.ToLower(i18n.GetEntityLabel(server, locale)), cleanedName) {
			serversFound = append(serversFound, server)
		}
	}

	return serversFound
}
