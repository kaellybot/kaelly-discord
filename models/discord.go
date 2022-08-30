package models

import (
	"github.com/bwmarrin/discordgo"
)

const (
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
