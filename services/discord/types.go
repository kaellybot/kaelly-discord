package discord

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

var (
	ErrInvalidInteractionType = errors.New("interaction type is not handled")
)

type Service interface {
	Listen() error
	IsConnected() bool
	Shutdown()
}

type Impl struct {
	session  *discordgo.Session
	commands []commands.DiscordCommand
}
