package constants

import (
	"github.com/bwmarrin/discordgo"
)

const (
	InvisibleCharacter       = "\u200b"
	MaxButtonPerActionRow    = 5
	MaxAlmanaxEffectPerEmbed = 5
	MaxBookRowPerEmbed       = 15
	MaxCharacterPerField     = 10
	MaxEquipmentPerField     = 8
	MaxIngredientsPerField   = 8

	MemberListLimit = 1000

	DefaultPage = 0

	MaxChoices = 25
)

const (
	DiscordCodeNotFound        = 0     // 404: Not Found.
	DiscordCodeTooManyWebhooks = 30007 // Maximum number of webhooks reached (15).
)

func GetIntents() discordgo.Intent {
	return discordgo.IntentMessageContent |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuilds |
		discordgo.IntentGuildMessages |
		discordgo.IntentGuildMessageReactions |
		discordgo.IntentGuildWebhooks
}
