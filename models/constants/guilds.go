package constants

import (
	"github.com/bwmarrin/discordgo"
)

type GuildConfig struct {
	Name            string
	Icon            string
	ServerID        string
	ChannelServers  []ChannelServer
	AlmanaxWebhooks []AlmanaxWebhook
	RssWebhooks     []RssWebhook
	TwitterWebhooks []TwitterWebhook
}

type ChannelServer struct {
	Channel  *discordgo.Channel
	ServerID string
}

type AlmanaxWebhook struct {
	Channel *discordgo.Channel
}

type RssWebhook struct {
	Channel *discordgo.Channel
	FeedID  string
}

type TwitterWebhook struct {
	Channel   *discordgo.Channel
	TwitterID string
}

