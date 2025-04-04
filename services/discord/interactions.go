package discord

import "github.com/bwmarrin/discordgo"

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
