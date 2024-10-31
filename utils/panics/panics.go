package panics

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

//nolint:nolintlint,exhaustive
func HandlePanic(session *discordgo.Session, event *discordgo.InteractionCreate) {
	r := recover()
	if r == nil {
		return
	}

	var commandName string
	switch event.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		commandName = event.ApplicationCommandData().Name
	case discordgo.InteractionMessageComponent:
		commandName = event.MessageComponentData().CustomID
	default:
		log.Warn().
			Uint32(constants.LogInteractionType, uint32(event.Interaction.Type)).
			Msgf("Cannot handle interaction type, continue recovering with this value as command Name")
		commandName = fmt.Sprintf("%v", event.Interaction.Type)
	}

	log.Error().Str(constants.LogCommand, commandName).
		Str(constants.LogPanic, fmt.Sprintf("%v", r)).
		Msgf("Panic occurred, sending an error message to user")

	content := i18n.Get(event.Locale, "panic")
	_, err := session.InteractionResponseEdit(event.Interaction, &discordgo.WebhookEdit{
		AllowedMentions: &discordgo.MessageAllowedMentions{},
		Attachments:     &[]*discordgo.MessageAttachment{},
		Components:      &[]discordgo.MessageComponent{},
		Content:         &content,
		Embeds:          &[]*discordgo.MessageEmbed{},
		Files:           []*discordgo.File{},
	})
	if err != nil {
		log.Warn().Err(err).Msgf("Could not respond to caller after panicking")
	}
}
