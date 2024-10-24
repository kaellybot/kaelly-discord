package config

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/streamers"
	"github.com/kaellybot/kaelly-discord/services/videasts"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	configurationRequestRoutingKey = "requests.configs"

	defaultIconSize = "128"
)

type Command struct {
	commands.AbstractCommand
	emojiService    emojis.Service
	feedService     feeds.Service
	guildService    guilds.Service
	serverService   servers.Service
	streamerService streamers.Service
	videastService  videasts.Service
	requestManager  requests.RequestManager
	handlers        commands.DiscordHandlers
}
