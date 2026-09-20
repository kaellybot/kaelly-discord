package commands

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/rs/zerolog/log"
)

func (command *AbstractCommand) CallHandler(s *discordgo.Session, i *discordgo.InteractionCreate,
	handlers DiscordHandlers) error {
	handler, found := handlers[i.Type]
	if !found {
		log.Error().
			Uint32(constants.LogInteractionType, uint32(i.Type)).
			Msgf("Interaction not handled, ignoring it")
		return nil
	}

	return handler(s, i)
}

func (command *AbstractCommand) HandleSubCommands(handlers SubCommandHandlers) DiscordHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if !IsApplicationCommand(i) {
			return nil
		}

		data := i.ApplicationCommandData()
		for _, subCommand := range data.Options {
			handler, found := handlers[subCommand.Name]
			if !found {
				return fmt.Errorf("%w: sub-command %q", ErrNoProvidedHandler, subCommand.Name)
			}

			if err := handler(s, i); err != nil {
				return err
			}
		}

		return nil
	}
}

func (command *AbstractCommand) HandleInteractionMessages(handlers InteractionMessageHandlers,
) DiscordHandler {
	// Iterated in a stable order: several patterns can match one custom ID, and
	// map iteration would pick a different handler from one call to the next.
	patterns := make([]*regexp.Regexp, 0, len(handlers))
	for regex := range handlers {
		patterns = append(patterns, regex)
	}
	slices.SortFunc(patterns, func(a, b *regexp.Regexp) int {
		return strings.Compare(a.String(), b.String())
	})

	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if !IsMessageCommand(i) {
			return nil
		}

		customID := i.MessageComponentData().CustomID
		for _, regex := range patterns {
			if regex.MatchString(customID) {
				return handlers[regex](s, i)
			}
		}

		return fmt.Errorf("%w: custom ID %q", ErrNoProvidedHandler, customID)
	}
}

func IsApplicationCommand(i *discordgo.InteractionCreate) bool {
	return i.Type == discordgo.InteractionApplicationCommand ||
		i.Type == discordgo.InteractionApplicationCommandAutocomplete
}

func IsMessageCommand(i *discordgo.InteractionCreate) bool {
	return i.Type == discordgo.InteractionMessageComponent
}
