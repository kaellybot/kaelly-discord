package discord

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func ExtractAPIError(err error) (*discordgo.APIErrorMessage, bool) {
	// Find JSON portion of the error
	startIndex := strings.Index(err.Error(), "{")
	if startIndex == -1 {
		return nil, false
	}

	// Parse JSON error into discordgo.APIErrorMessage
	var discordError discordgo.APIErrorMessage
	jsonPart := err.Error()[startIndex:]
	if errJSON := discordgo.Unmarshal([]byte(jsonPart), &discordError); errJSON != nil {
		return nil, false
	}

	return &discordError, true
}

// AnswerError tells the caller what went wrong, in their language, and records
// it for us. It is the only place mapping a failure to a message: handlers
// return their errors instead of answering, so the wording stays consistent.
//
// The category drives both the message and the log level: a stale component or
// an unreachable backend is routine, anything else is a bug worth alerting on.
func AnswerError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	var key string
	switch {
	case errors.Is(err, commands.ErrStaleInteraction):
		key = "errors.unknown_interaction"
		log.Warn().Err(err).
			Str(constants.LogCommand, interactionName(i)).
			Msgf("Interaction cannot be interpreted, telling the caller")
	case errors.Is(err, commands.ErrUnavailable):
		key = "errors.unavailable"
		log.Warn().Err(err).
			Str(constants.LogCommand, interactionName(i)).
			Msgf("Dependency unavailable, telling the caller")
	default:
		key = "panic"
		log.Error().Err(err).
			Str(constants.LogCommand, interactionName(i)).
			Msgf("Interaction failed, telling the caller")
	}

	// Only a deferred interaction has an answer to edit: an autocomplete, a
	// ping or a modal submit would reject the edit, and the failure is logged
	// above either way.
	if !isDeferred(i) {
		return
	}

	content := i18n.Get(i.Locale, key)
	if _, errEdit := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); errEdit != nil {
		log.Warn().Err(errEdit).Msgf("Could not tell the caller their interaction failed")
	}
}

//nolint:exhaustive // the types left are the ones never deferred.
func isDeferred(i *discordgo.InteractionCreate) bool {
	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionMessageComponent:
		return true
	default:
		return false
	}
}

// interactionName switches on the payload rather than on the interaction type:
// discordgo's accessors assert on it, and the error path is the last place
// that should panic.
func interactionName(i *discordgo.InteractionCreate) string {
	switch data := i.Data.(type) {
	case discordgo.ApplicationCommandInteractionData:
		return data.Name
	case discordgo.MessageComponentInteractionData:
		return data.CustomID
	default:
		return i.Type.String()
	}
}
