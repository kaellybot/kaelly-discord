package commands

import "github.com/kaellybot/kaelly-discord/models/constants"

type Command interface {
	GetDiscordCommand() *constants.DiscordCommand
}
