package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (service *Impl) deferInteraction(session *discordgo.Session,
	event *discordgo.InteractionCreate) error {
	switch event.Type {
	case discordgo.InteractionApplicationCommand:
		return session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
	case discordgo.InteractionMessageComponent:
		return session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
	case discordgo.InteractionPing, discordgo.InteractionApplicationCommandAutocomplete, discordgo.InteractionModalSubmit:
		fallthrough
	default:
	}

	return nil
}

// answerUnmatchedInteraction warns the caller instead of leaving their
// interaction in its deferred state until Discord times it out.
func (service *Impl) answerUnmatchedInteraction(session *discordgo.Session,
	event *discordgo.InteractionCreate) {
	switch event.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionMessageComponent:
	case discordgo.InteractionPing, discordgo.InteractionApplicationCommandAutocomplete,
		discordgo.InteractionModalSubmit:
		fallthrough
	default:
		// Never deferred, so there is no answer to edit.
		return
	}

	log.Warn().
		Uint32(constants.LogInteractionType, uint32(event.Type)).
		Msgf("Interaction handled by no command, answering an error message to caller")

	content := i18n.Get(event.Locale, "errors.unknown_interaction")
	_, err := session.InteractionResponseEdit(event.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		log.Warn().Err(err).Msgf("Could not respond to caller after unmatched interaction")
	}
}
