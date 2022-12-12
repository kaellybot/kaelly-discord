package servers

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/servers"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type ServerService interface {
	GetServer(id string) (entities.Server, bool)
	GetServers() []entities.Server
	FindServers(name string, locale discordgo.Locale) []entities.Server
}

type ServerServiceImpl struct {
	transformer transform.Transformer
	serversMap  map[string]entities.Server
	servers     []entities.Server
	repository  repository.ServerRepository
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
		transformer: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		serversMap:  serversMap,
		servers:     servers,
		repository:  repository,
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
	cleanedName, _, err := transform.String(service.transformer, strings.ToLower(name))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize server name, returning empty server list")
		return serversFound
	}

	for _, server := range service.servers {
		currentCleanedName, _, err := transform.String(service.transformer, strings.ToLower(translators.GetEntityLabel(server, locale)))
		if err == nil && strings.HasPrefix(currentCleanedName, cleanedName) {
			serversFound = append(serversFound, server)
		}
	}

	return serversFound
}
