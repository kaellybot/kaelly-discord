package job

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

func (command *JobCommand) displayRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	// TODO Check in DB
}
