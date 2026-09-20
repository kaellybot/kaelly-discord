package discord

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

func GetUserID(i *discordgo.Interaction) string {
	if i == nil {
		return ""
	}

	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}

	if i.User != nil {
		return i.User.ID
	}

	return ""
}

func GetMemberNickNames(s *discordgo.Session, guildID string) (map[string]any, error) {
	members, err := s.GuildMembers(guildID, "", constants.MemberListLimit)
	if err != nil {
		return nil, err
	}

	properties := make(map[string]any)
	for _, member := range members {
		username := member.Nick
		if len(username) == 0 {
			username = member.User.Username
		}
		properties[member.User.ID] = username
	}

	return properties, nil
}

func GetInt64Value(data discordgo.MessageComponentInteractionData) (int64, error) {
	values := data.Values
	if len(values) != 1 {
		return 0, fmt.Errorf("%w: %v values selected instead of one",
			commands.ErrInvalidInteraction, len(values))
	}

	value, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		// Classified here so every select menu reports it the same way: the
		// value comes from the caller's client, it is not our logic failing.
		return 0, fmt.Errorf("%w: %w", commands.ErrInvalidInteraction, err)
	}

	return value, nil
}
