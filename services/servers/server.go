package servers

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
			{Name: "hellmina"},
			{Name: "draconiros"},
			{Name: "imagiro"},
			{Name: "orukam"},
			{Name: "ombre"},
			{Name: "talKasha"},
			{Name: "tylezia"},
		},
	}, nil
}

func (service *ServerServiceImpl) GetServers() []models.Server {
	return service.servers
}

func (service *ServerServiceImpl) FindServers(name string, locale discordgo.Locale) []models.Server {
	serversFound := make([]models.Server, 0)
	cleanedName := strings.ToLower(name)

	for _, server := range service.servers {
		if strings.HasPrefix(strings.ToLower(server.Name), cleanedName) {
			serversFound = append(serversFound, server)
		}
	}

	return serversFound
}
