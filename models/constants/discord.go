package constants

import (
	"github.com/bwmarrin/discordgo"
)

type DiscordHandlers map[discordgo.InteractionType]DiscordHandler
type DiscordHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale)

type DiscordCommand struct {
	Identity discordgo.ApplicationCommand
	Handlers DiscordHandlers
}

func GetDMPermission() *bool {
	var dmPermission = false
	return &dmPermission
}

func GetDefaultPermission() *int64 {
	var defaultPermission int64 = discordgo.PermissionViewChannel
	return &defaultPermission
}

func GetManageServerPermission() *int64 {
	var manageServerPermission int64 = discordgo.PermissionManageServer
	return &manageServerPermission
}
func GetIntents() discordgo.Intent {
	return discordgo.IntentMessageContent |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuilds |
		discordgo.IntentGuildMessages |
		discordgo.IntentGuildMessageReactions |
		discordgo.IntentGuildWebhooks
}
