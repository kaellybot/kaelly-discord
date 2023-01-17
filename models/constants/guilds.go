package constants

import (
	"github.com/bwmarrin/discordgo"
)

type GuildConfig struct {
	Name            string
	Icon            string
	ServerId        string
	ChannelServers  []ChannelServer
	ChannelWebhooks []ChannelWebhook
}

type ChannelServer struct {
	Channel  *discordgo.Channel
	ServerId string
}

type ChannelWebhook struct {
	Channel *discordgo.Channel
	// TODO
}
