package panics

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func HandlePanic(session *discordgo.Session, event *discordgo.InteractionCreate) {
	r := recover()
	if r == nil {
		return
	}

	// TODO not always ApplicationCommandData
	log.Error().Str(constants.LogCommand, event.ApplicationCommandData().Name).Str(constants.LogPanic, fmt.Sprintf("%v", r)).Msgf("Panic occurred, sending an error message to user")
	err := session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: i18n.Get(event.Locale, "panic"),
		},
	})

	if err != nil {
		if interactionHasAlreadyBeenAcknowledged(err) {
			content := i18n.Get(event.Locale, "panic")
			_, err = session.InteractionResponseEdit(event.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			if err != nil {
				log.Warn().Err(err).Msgf("Could not respond to caller after panicking")
			}
		} else {
			log.Warn().Err(err).Msgf("Could not respond to caller after panicking")
		}
	}
}

func interactionHasAlreadyBeenAcknowledged(err error) bool {
	restErr, ok := err.(*discordgo.RESTError)
	return ok && restErr != nil &&
		restErr.Message != nil &&
		restErr.Message.Code == discordgo.ErrCodeInteractionHasAlreadyBeenAcknowledged
}
