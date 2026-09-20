package commands

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

var (
	// ErrStaleInteraction marks an interaction that cannot be interpreted any
	// more: a custom ID crafted by an older version of the bot, a component
	// nothing handles. The caller triggered it, it is not a bug.
	ErrStaleInteraction = errors.New("interaction cannot be interpreted")

	// ErrUnavailable marks a dependency that refused to answer: the broker, the
	// database, a backend service. Trying again later usually works.
	ErrUnavailable = errors.New("dependency is unavailable")

	ErrInvalidAnswerMessage = fmt.Errorf("answer message is not valid: %w", ErrUnavailable)
	ErrNoProvidedHandler    = fmt.Errorf("no handler provided: %w", ErrStaleInteraction)
	ErrInvalidInteraction   = fmt.Errorf("message interaction is not valid: %w", ErrStaleInteraction)

	// ErrRequestPropertyNotFound is wiring gone wrong between the request and
	// its answer: a bug, so it stays in the default category.
	ErrRequestPropertyNotFound = errors.New("request property is not found")
)

type DiscordCommand interface {
	GetName() string
	GetDescriptions(lg discordgo.Locale) []Description
	Matches(i *discordgo.InteractionCreate) bool
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

type AbstractCommand struct {
	DiscordID string
}

type Description struct {
	Name        string
	CommandID   string
	Description string
	TutorialURL string
}

// DiscordHandler answers an interaction. Returning an error means the caller
// could not be served: the router turns it into a message for them, so a
// handler never answers an error itself.
type DiscordHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) error
type DiscordHandlers map[discordgo.InteractionType]DiscordHandler
type SubCommandHandlers map[string]DiscordHandler
type InteractionMessageHandlers map[*regexp.Regexp]DiscordHandler
