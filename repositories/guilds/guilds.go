package guilds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetServer(guildID, channelID string) (*entities.Server, error) {
	var serverIDs []string
	err := repo.db.GetDB().Table("guilds").
		Select("COALESCE(channel_servers.server_id, guilds.server_id) as server").
		Joins("LEFT JOIN channel_servers ON guilds.id = channel_servers.guild_id AND channel_servers.channel_id = ?", channelID).
		Where("guilds.id = ?", guildID).
		Pluck("server", &serverIDs).Error
	if err != nil {
		return nil, err
	}
	if len(serverIDs) == 0 {
		return nil, nil
	}

	return &entities.Server{ID: serverIDs[0]}, nil
}
