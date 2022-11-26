package commands

import "github.com/kaellybot/kaelly-discord/models"

type Command interface {
	GetDiscordCommand() *models.DiscordCommand
}
