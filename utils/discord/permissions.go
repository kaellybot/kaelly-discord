package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func HasPermissions(s *discordgo.Session, channelID string, requiredPermissions int64) bool {
	perms, errPerm := s.UserChannelPermissions(s.State.User.ID, channelID)
	if errPerm != nil {
		log.Warn().Err(errPerm).
			Msgf("Cannot retrieve channel permissions, returning false")
		return false
	}
	return perms&requiredPermissions == requiredPermissions
}
