package commands

import (
	"errors"

	"github.com/kaellybot/kaelly-discord/models/constants"
)

var (
	ErrInvalidAnswerMessage = errors.New("Answer message is not valid")
)

type Command interface {
	GetDiscordCommand() *constants.DiscordCommand
}
