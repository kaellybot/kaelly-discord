package help

import "github.com/kaellybot/kaelly-discord/commands"

type Command struct {
	commands.AbstractCommand
	commands *[]commands.DiscordCommand
	handlers commands.DiscordHandlers
}

const (
	menuCommandName = "menu"
)
