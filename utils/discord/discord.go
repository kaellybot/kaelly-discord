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

// GetTargetMember returns the member whose books are requested: the target of
// a user command, else the member picked through the optionName option of its
// slash twin, else the author. The resolved user is attached to the member.
// An option also accepts the ID of a user who is not in the guild: they come
// back without guild data, where a target is always expected to be a member.
func GetTargetMember(i *discordgo.InteractionCreate, optionName string) (*discordgo.Member, error) {
	data := i.ApplicationCommandData()
	userID := data.TargetID
	if len(userID) == 0 {
		userID = getUserOptionValue(optionName, data.Options)
	}

	if len(userID) == 0 {
		if i.Member == nil || i.Member.User == nil {
			return nil, fmt.Errorf("%w: no author to fall back on", commands.ErrInvalidInteraction)
		}

		return i.Member, nil
	}

	if data.Resolved == nil {
		return nil, fmt.Errorf("%w: user %q is not resolved", commands.ErrInvalidInteraction, userID)
	}

	user, found := data.Resolved.Users[userID]
	if !found {
		return nil, fmt.Errorf("%w: user %q is not resolved", commands.ErrInvalidInteraction, userID)
	}

	member, found := data.Resolved.Members[userID]
	if !found {
		if len(data.TargetID) > 0 {
			return nil, fmt.Errorf("%w: member %q is not resolved", commands.ErrInvalidInteraction, userID)
		}

		member = &discordgo.Member{}
	}
	member.User = user

	return member, nil
}

func getUserOptionValue(optionName string, options []*discordgo.ApplicationCommandInteractionDataOption) string {
	for _, option := range options {
		if option.Type == discordgo.ApplicationCommandOptionUser && option.Name == optionName {
			return option.UserValue(nil).ID
		}

		if userID := getUserOptionValue(optionName, option.Options); len(userID) > 0 {
			return userID
		}
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
