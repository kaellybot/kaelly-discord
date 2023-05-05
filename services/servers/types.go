package servers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/servers"
	"golang.org/x/text/transform"
)

type Service interface {
	GetServer(id string) (entities.Server, bool)
	GetServers() []entities.Server
	FindServers(name string, locale discordgo.Locale) []entities.Server
}

type Impl struct {
	transformer transform.Transformer
	serversMap  map[string]entities.Server
	servers     []entities.Server
	repository  repository.Repository
}
