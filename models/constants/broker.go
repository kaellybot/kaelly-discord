package constants

import (
	"github.com/spf13/viper"
)

const (
	AboutRoutingKey                = "requests.about"
	AlignRequestRoutingKey         = "requests.books"
	AlmanaxRequestRoutingKey       = "requests.encyclopedias"
	CompetitionRequestRoutingKey   = "requests.competitions"
	ConfigurationRequestRoutingKey = "requests.configs"
	HelpRoutingKey                 = "requests.help"
	ItemRequestRoutingKey          = "requests.encyclopedias"
	JobRequestRoutingKey           = "requests.books"
	PortalRequestRoutingKey        = "requests.portals"
	SetRequestRoutingKey           = "requests.encyclopedias"

	GuildNewsRoutingKey = "news.guild"
)

func GetRabbitMQClientID() string {
	return Name + "-" + viper.GetString(ShardID)
}
