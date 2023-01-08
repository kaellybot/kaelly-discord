package servers

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type GuildRepository interface {
	GetServer(guildID, channelId string) (*entities.Server, error)
}

type GuildRepositoryImpl struct {
	db databases.MySQLConnection
}

func New(db databases.MySQLConnection) *GuildRepositoryImpl {
	return &GuildRepositoryImpl{db: db}
}

func (repo *GuildRepositoryImpl) GetServer(guildId, channelId string) (*entities.Server, error) {
	var serverId string
	err := repo.db.GetDB().Table("guild").
		Select("COALESCE(channel_servers.server_id, guild.server_id) as server").
		Joins("LEFT JOIN channel_servers ON guild.id = channel_servers.guild_id AND channel_servers.channel_id = ?", channelId).
		Where("guild.id = ?", guildId).
		First(&serverId).Error
	if err != nil {
		return nil, err
	}
	if len(serverId) == 0 {
		return nil, nil
	}

	return &entities.Server{Id: serverId}, nil
}
