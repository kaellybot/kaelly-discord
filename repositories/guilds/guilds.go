package guilds

import (
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

// Exists is scoped to this bot's game, like GetServers: a guild registered by the
// other game's bot is not this bot's guild, and treating it as known would suppress
// this bot's create/delete news and stop the configurator creating its row.
func (repo *Impl) Exists(guildID string) (bool, error) {
	// gorm drops zero-valued fields from a struct condition, so an empty ID
	// would match any guild: no ID, no guild.
	if guildID == "" {
		return false, nil
	}

	guild := entities.Guild{ID: guildID, Game: constants.GetGame().AMQPGame}
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
	game := constants.GetGame().AMQPGame
	err := repo.db.GetDB().Table("guilds").
		Select("COALESCE(channel_servers.server_id, guilds.server_id, '') as server").
		Joins("LEFT JOIN channel_servers ON guilds.id = channel_servers.guild_id "+
			"AND channel_servers.channel_id = ? AND channel_servers.game = guilds.game", channelID).
		Where("guilds.id = ? AND guilds.game = ?", guildID, game).
		Pluck("server", &serverIDs).Error
	if err != nil {
		return entities.Server{}, false, err
	}
	if len(serverIDs) == 0 || serverIDs[0] == "" {
		return entities.Server{}, false, nil
	}

	return entities.Server{ID: serverIDs[0], Game: game}, true, nil
}
