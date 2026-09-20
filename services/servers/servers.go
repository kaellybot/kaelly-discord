package servers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/servers"

	"github.com/kaellybot/kaelly-discord/utils/translators"
	"github.com/rs/zerolog/log"
)

func New(repository repository.Repository) (*Impl, error) {
	servers, err := repository.GetServers()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(servers)).
		Msgf("Servers loaded")

	serversMap := make(map[string]entities.Server)
	for _, server := range servers {
		serversMap[server.ID] = server
	}

	return &Impl{
		serversMap: serversMap,
		servers:    servers,
		repository: repository,
	}, nil
}

func (service *Impl) GetServers() []entities.Server {
	return service.servers
}

func (service *Impl) GetServer(id string) (entities.Server, bool) {
	server, found := service.serversMap[id]
	return server, found
}

func (service *Impl) FindServers(name string, locale discordgo.Locale, limit int) []entities.Server {
	serversFound := make([]entities.Server, 0)
	cleanedName, err := translators.NormalizeName(name)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot normalize server name, returning empty server list")
		return serversFound
	}

	for _, server := range service.servers {
		currentCleanedName, errStr := translators.NormalizeName(translators.GetEntityLabel(server, locale))
		if errStr == nil && strings.Contains(currentCleanedName, cleanedName) {
			if currentCleanedName == cleanedName {
				return []entities.Server{server}
			}

			serversFound = append(serversFound, server)
		}
	}

	if len(serversFound) > limit {
		return serversFound[:limit]
	}

	return serversFound
}
