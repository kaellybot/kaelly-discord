package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type Service interface {
	Listen() error
	RegisterCommands() error
	Shutdown() error
}

type Impl struct {
	session  *discordgo.Session
	commands []*constants.DiscordCommand
}
