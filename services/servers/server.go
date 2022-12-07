package servers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
)

type ServerService interface {
	GetServers() []entities.Server
	FindServers(name string, locale discordgo.Locale) []entities.Server
}

type ServerServiceImpl struct {
	servers []entities.Server
}

func New() (*ServerServiceImpl, error) {
	return &ServerServiceImpl{
		servers: []entities.Server{
			{Id: "hell mina"},
			{Id: "draconiros"},
			{Id: "imagiro"},
			{Id: "orukam"},
			{Id: "ombre"},
			{Id: "talKasha"},
			{Id: "tylezia"},
		},
	}, nil
}

func (service *ServerServiceImpl) GetServers() []entities.Server {
	return service.servers
}

func (service *ServerServiceImpl) FindServers(name string, locale discordgo.Locale) []entities.Server {
	serversFound := make([]entities.Server, 0)
	cleanedName := strings.ToLower(name)

	for _, server := range service.servers {
		if strings.HasPrefix(strings.ToLower(server.Id), cleanedName) {
			serversFound = append(serversFound, server)
		}
	}

	return serversFound
}
