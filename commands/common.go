package commands

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrInvalidAnswerMessage = errors.New("Answer message is not valid")
)

func DeferInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}
