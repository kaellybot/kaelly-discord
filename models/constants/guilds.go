package constants

type GuildConfig struct {
	Name            string
	Icon            string
	ChannelServers  []ChannelServer
	ChannelWebhooks []ChannelWebhook
}

type ChannelServer struct {
	ChannelName string
	ServerId    string
}

type ChannelWebhook struct {
	ChannelName string
	// TODO
}
