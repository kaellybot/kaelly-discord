package commands

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrInvalidAnswerMessage = errors.New("answer message is not valid")
)

type DiscordCommand interface {
	Matches(i *discordgo.InteractionCreate) bool
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale)
}

type AbstractCommand struct {}

type DiscordHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale)
type DiscordHandlers map[discordgo.InteractionType]DiscordHandler
