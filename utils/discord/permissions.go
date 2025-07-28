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

	// By default, channels permissions overrides guild permissions.
	// Administrator role is the only permission above the previous rule:
	// having it give you the right no matter channels permissions.
	// By comparing to the required permissions, it verifies that all bits in requiredPermissions are set;
	// Comparing by 0 is to check if at least one of the compared permission is present.
	return perms&discordgo.PermissionAdministrator != 0 ||
		perms&requiredPermissions == requiredPermissions
}
