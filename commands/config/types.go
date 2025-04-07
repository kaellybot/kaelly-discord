package config

import (
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/services/almanaxes"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/twitters"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

const (
	defaultIconSize = "128"
)

type Command struct {
	commands.AbstractCommand
	almanaxService almanaxes.Service
	emojiService   emojis.Service
	feedService    feeds.Service
	guildService   guilds.Service
	serverService  servers.Service
	twitterService twitters.Service
	requestManager requests.RequestManager
	handlers       commands.DiscordHandlers
}
