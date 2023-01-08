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
	var serverIds []string
	err := repo.db.GetDB().Table("guilds").
		Select("COALESCE(channel_servers.server_id, guilds.server_id) as server").
		Joins("LEFT JOIN channel_servers ON guilds.id = channel_servers.guild_id AND channel_servers.channel_id = ?", channelId).
		Where("guilds.id = ?", guildId).
		Pluck("server", &serverIds).Error
	if err != nil {
		return nil, err
	}
	if len(serverIds) == 0 {
		return nil, nil
	}

	return &entities.Server{Id: serverIds[0]}, nil
}
