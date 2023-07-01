package commands

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrInvalidAnswerMessage    = errors.New("answer message is not valid")
	ErrNoSubCommandHandler     = errors.New("no sub command handler provided")
	ErrInvalidInteraction      = errors.New("message interaction is not valid")
	ErrRequestPropertyNotFound = errors.New("request property is not found")
)

type DiscordCommand interface {
	Matches(i *discordgo.InteractionCreate) bool
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale)
}

type AbstractCommand struct{}

type DiscordHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale)
type DiscordHandlers map[discordgo.InteractionType]DiscordHandler
type SubCommandHandlers map[string]DiscordHandler
