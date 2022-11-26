package models

import (
	"github.com/bwmarrin/discordgo"
)

type DiscordHandlers map[discordgo.InteractionType]DiscordHandler
type DiscordHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

type DiscordCommand struct {
	Identity discordgo.ApplicationCommand
	Handlers DiscordHandlers
}

var (
	DMPermission = false

	DefaultPermission int64 = discordgo.PermissionSendMessages |
		discordgo.PermissionEmbedLinks |
		discordgo.PermissionAttachFiles |
		discordgo.PermissionUseExternalEmojis |
		discordgo.PermissionSendMessagesInThreads

	Intents discordgo.Intent = discordgo.IntentMessageContent |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuilds |
		discordgo.IntentGuildMessages |
		discordgo.IntentGuildMessageReactions
)
