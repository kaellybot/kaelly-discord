package models

import (
	"github.com/bwmarrin/discordgo"
)

type DiscordCommand struct {
	Name            string
	CommandType     discordgo.ApplicationCommandType
	InteractionType discordgo.InteractionType
	Handler         func(session *discordgo.Session, interaction *discordgo.InteractionCreate)
}

func (cmd *DiscordCommand) GetApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name: cmd.Name,
		Type: cmd.CommandType,
	}
}

const (
	Intents discordgo.Intent = discordgo.IntentMessageContent |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuilds |
		discordgo.IntentGuildMessages |
		discordgo.IntentGuildMessageReactions
)
