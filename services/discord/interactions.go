package discord

import "github.com/bwmarrin/discordgo"

func (service *Impl) deferInteraction(session *discordgo.Session,
	event *discordgo.InteractionCreate) error {
	if event.Interaction.Type == discordgo.InteractionApplicationCommand {
		return session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
	} else if event.Interaction.Type == discordgo.InteractionMessageComponent {
		return session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
	}

	return nil
}
