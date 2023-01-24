package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

type DiscordService interface {
	Listen() error
	RegisterCommands() error
	Shutdown() error
}

type DiscordServiceImpl struct {
	session  *discordgo.Session
	commands []*constants.DiscordCommand
}
