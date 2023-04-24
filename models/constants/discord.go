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

var (
	DMPermission = false

	DefaultPermission      int64 = discordgo.PermissionViewChannel
	ManageServerPermission int64 = discordgo.PermissionManageServer

	Intents discordgo.Intent = discordgo.IntentMessageContent |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuilds |
		discordgo.IntentGuildMessages |
		discordgo.IntentGuildMessageReactions |
		discordgo.IntentGuildWebhooks
)
