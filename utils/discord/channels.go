package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/rs/zerolog/log"
)

// ResolveParentChannelID returns the channel answering for channelID: a thread
// owns neither permission overwrites nor a configuration of its own, so its
// parent answers for both. Any other channel, and any channel we cannot look
// up, answers for itself.
func ResolveParentChannelID(s *discordgo.Session, channelID string) string {
	if channelID == "" {
		return channelID
	}

	channel, err := s.State.Channel(channelID)
	if err != nil || channel == nil {
		channel, err = s.Channel(channelID)
		if err != nil {
			log.Warn().Err(err).
				Str(constants.LogChannelID, channelID).
				Msgf("Cannot retrieve channel, continuing with it as its own parent")
			return channelID
		}
	}

	if channel.IsThread() && channel.ParentID != "" {
		return channel.ParentID
	}

	return channelID
}
