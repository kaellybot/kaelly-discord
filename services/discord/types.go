package discord

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

var (
	ErrInvalidInteractionType = errors.New("Interaction type is not handled")
)

type Service interface {
	Listen() error
	Shutdown()
}

type Impl struct {
	session  *discordgo.Session
	commands []commands.DiscordCommand
}
