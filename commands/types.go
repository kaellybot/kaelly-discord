package commands

import (
	"errors"

	"github.com/kaellybot/kaelly-discord/models/constants"
)

var (
	ErrInvalidAnswerMessage = errors.New("answer message is not valid")
)

type SlashCommand interface {
	GetSlashCommand() *constants.DiscordCommand
}

type UserCommand interface {
	GetUserCommand() *constants.DiscordCommand
}
