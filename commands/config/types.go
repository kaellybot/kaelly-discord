package config

import (
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	commandName           = "config"
	getSubCommandName     = "get"
	almanaxSubCommandName = "almanax"
	rssSubCommandName     = "rss"
	twitterSubCommandName = "twitter"
	serverSubCommandName  = "server"

	serverOptionName   = "server"
	channelOptionName  = "channel"
	feedTypeOptionName = "type"
	languageOptionName = "language"
	enabledOptionName  = "enabled"

	configurationRequestRoutingKey = "requests.configs"
)

type ConfigCommand struct {
	guildService   guilds.GuildService
	feedService    feeds.FeedService
	serverService  servers.ServerService
	requestManager requests.RequestManager
}
