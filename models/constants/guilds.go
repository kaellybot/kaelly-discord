package constants

import (
	"github.com/bwmarrin/discordgo"
)

type GuildConfig struct {
	Name             string
	Icon             string
	ServerID         string
	ServerChannels   []ServerChannel
	NotifiedChannels []NotifiedChannel
}

type ServerChannel struct {
	Channel  *discordgo.Channel
	ServerID string
}

type NotifiedChannel struct {
	Channel *discordgo.Channel
	//  TODO
}
