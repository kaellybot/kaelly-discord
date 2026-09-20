package mappers

import "github.com/bwmarrin/discordgo"

// mapMemberIdentity returns the name and the avatar to display for a member,
// tolerating a missing user payload.
func mapMemberIdentity(member *discordgo.Member) (string, string) {
	if member.User == nil {
		return member.Nick, ""
	}

	userName := member.Nick
	if len(userName) == 0 {
		userName = member.User.Username
	}

	// discordgo already falls back to the user avatar when the member has none.
	return userName, member.AvatarURL("")
}
