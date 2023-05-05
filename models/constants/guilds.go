package constants

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
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

type ChannelWebhook struct {
	Channel *discordgo.Channel
	Locale  amqp.Language
}

type AlmanaxWebhook struct {
	ChannelWebhook
}

type RssWebhook struct {
	ChannelWebhook
	FeedID string
}

type TwitterWebhook struct {
	ChannelWebhook
	TwitterName string
}
