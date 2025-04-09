package guilds

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) Exists(guildID string) (bool, error) {
	guild := entities.Guild{ID: guildID}
	var count int64
	err := repo.db.GetDB().
		Model(&guild).
		Where(&guild).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *Impl) GetServer(guildID, channelID string) (entities.Server, bool, error) {
	var serverIDs []string
	err := repo.db.GetDB().Table("guilds").
		Select("COALESCE(channel_servers.server_id, guilds.server_id, '') as server").
		Joins("LEFT JOIN channel_servers ON guilds.id = channel_servers.guild_id "+
			"AND channel_servers.channel_id = ?", channelID).
		Where("guilds.id = ?", guildID).
		Pluck("server", &serverIDs).Error
	if err != nil {
		return entities.Server{}, false, err
	}
	if len(serverIDs) == 0 || serverIDs[0] == "" {
		return entities.Server{}, false, nil
	}

	return entities.Server{ID: serverIDs[0]}, true, nil
}
