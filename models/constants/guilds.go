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
	TwitchWebhooks  []TwitchWebhook
	TwitterWebhooks []TwitterWebhook
	YoutubeWebhooks []YoutubeWebhook
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

type TwitchWebhook struct {
	ChannelWebhook
	StreamerID string
}

type TwitterWebhook struct {
	ChannelWebhook
	TwitterID string
}

type YoutubeWebhook struct {
	ChannelWebhook
	VideastID string
}
