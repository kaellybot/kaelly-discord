package constants

import (
	"github.com/bwmarrin/discordgo"
)

const (
	InvisibleCharacter = "\u200b"
	MaxButtonPerActionRow = 5
)

func GetIntents() discordgo.Intent {
	return discordgo.IntentMessageContent |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuilds |
		discordgo.IntentGuildMessages |
		discordgo.IntentGuildMessageReactions |
		discordgo.IntentGuildWebhooks
}
