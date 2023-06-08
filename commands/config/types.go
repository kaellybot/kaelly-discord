package config

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	configurationRequestRoutingKey = "requests.configs"

	defaultIconSize = "128"
)

type Command struct {
	commands.AbstractCommand
	guildService   guilds.Service
	feedService    feeds.Service
	serverService  servers.Service
	requestManager requests.RequestManager
	handlers       commands.DiscordHandlers
}
