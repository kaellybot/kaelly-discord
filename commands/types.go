package commands

import (
	"errors"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrInvalidAnswerMessage    = errors.New("answer message is not valid")
	ErrNoProvidedHandler       = errors.New("no handler provided")
	ErrInvalidInteraction      = errors.New("message interaction is not valid")
	ErrRequestPropertyNotFound = errors.New("request property is not found")
)

type DiscordCommand interface {
	GetName() string
	GetDescriptions(lg discordgo.Locale) []Description
	Matches(i *discordgo.InteractionCreate) bool
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type AbstractCommand struct{}

type Description struct {
	Name        string
	CommandID   string
	Description string
	TutorialURL string
}

type DiscordHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)
type DiscordHandlers map[discordgo.InteractionType]DiscordHandler
type SubCommandHandlers map[string]DiscordHandler
type InteractionMessageHandlers map[*regexp.Regexp]DiscordHandler
