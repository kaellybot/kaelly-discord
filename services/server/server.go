package server

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
)

type ServerService interface {
	GetServers() []models.Server
	FindServers(name string, locale discordgo.Locale) []models.Server
}

type ServerServiceImpl struct {
	servers []models.Server
}

func New() (*ServerServiceImpl, error) {
	return &ServerServiceImpl{
		servers: []models.Server{
			{Name: "Hell Mina"},
			{Name: "Draconiros"},
			{Name: "Imagiro"},
			{Name: "Orukam"},
			{Name: "Ombre"},
			{Name: "Tal Kasha"},
			{Name: "Tylezia"},
		},
	}, nil
}

func (service *ServerServiceImpl) GetServers() []models.Server {
	return service.servers
}

func (service *ServerServiceImpl) FindServers(name string, locale discordgo.Locale) []models.Server {
	serversFound := make([]models.Server, 0)

	for _, server := range service.servers {
		if strings.HasPrefix(server.Name, name) {
			serversFound = append(serversFound, server)
		}
	}

	return serversFound
}
