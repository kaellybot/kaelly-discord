package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const memberOption = "member"

func targetInteraction(data discordgo.ApplicationCommandInteractionData) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		Type: discordgo.InteractionApplicationCommand, Data: data,
		Member: &discordgo.Member{Nick: "Author", User: &discordgo.User{ID: "author"}},
	}}
}

func subCommandWith(options ...*discordgo.ApplicationCommandInteractionDataOption,
) []*discordgo.ApplicationCommandInteractionDataOption {
	return []*discordgo.ApplicationCommandInteractionDataOption{{
		Name: "user", Type: discordgo.ApplicationCommandOptionSubCommand, Options: options,
	}}
}

func memberValue(userID string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name: memberOption, Type: discordgo.ApplicationCommandOptionUser, Value: userID,
	}
}

func resolved(withMember bool) *discordgo.ApplicationCommandInteractionDataResolved {
	data := &discordgo.ApplicationCommandInteractionDataResolved{
		Users:   map[string]*discordgo.User{"u1": {ID: "u1", Username: "alice"}},
		Members: map[string]*discordgo.Member{},
	}
	if withMember {
		data.Members["u1"] = &discordgo.Member{Nick: "Ally"}
	}
	return data
}

func TestGetTargetMember_UserCommandTargetWins(t *testing.T) {
	member, err := GetTargetMember(targetInteraction(discordgo.ApplicationCommandInteractionData{
		TargetID: "u1", Resolved: resolved(true),
	}), memberOption)

	require.NoError(t, err)
	assert.Equal(t, "Ally", member.Nick)
	require.NotNil(t, member.User, "resolved user is attached to the member")
	assert.Equal(t, "alice", member.User.Username)
}

func TestGetTargetMember_SlashOptionIsResolved(t *testing.T) {
	member, err := GetTargetMember(targetInteraction(discordgo.ApplicationCommandInteractionData{
		Options: subCommandWith(memberValue("u1")), Resolved: resolved(true),
	}), memberOption)

	require.NoError(t, err)
	assert.Equal(t, "Ally", member.Nick)
	assert.Equal(t, "u1", member.User.ID)
}

func TestGetTargetMember_SlashOptionAcceptsUserOutsideGuild(t *testing.T) {
	member, err := GetTargetMember(targetInteraction(discordgo.ApplicationCommandInteractionData{
		Options: subCommandWith(memberValue("u1")), Resolved: resolved(false),
	}), memberOption)

	require.NoError(t, err)
	assert.Empty(t, member.Nick)
	assert.Equal(t, "alice", member.User.Username)
}

func TestGetTargetMember_AuthorIsTheDefault(t *testing.T) {
	for name, options := range map[string][]*discordgo.ApplicationCommandInteractionDataOption{
		"no option": subCommandWith(),
		"another option": subCommandWith(&discordgo.ApplicationCommandInteractionDataOption{
			Name: "server", Type: discordgo.ApplicationCommandOptionString, Value: "imagiro",
		}),
	} {
		t.Run(name, func(t *testing.T) {
			member, err := GetTargetMember(targetInteraction(discordgo.ApplicationCommandInteractionData{
				Options: options,
			}), memberOption)

			require.NoError(t, err)
			assert.Equal(t, "Author", member.Nick)
			assert.Equal(t, "author", member.User.ID)
		})
	}
}

func TestGetTargetMember_UnreadableMembersAreStale(t *testing.T) {
	withoutAuthor := targetInteraction(discordgo.ApplicationCommandInteractionData{})
	withoutAuthor.Member = nil

	cases := map[string]*discordgo.InteractionCreate{
		"target without resolved data": targetInteraction(discordgo.ApplicationCommandInteractionData{
			TargetID: "u1",
		}),
		"target without resolved member": targetInteraction(discordgo.ApplicationCommandInteractionData{
			TargetID: "u1", Resolved: resolved(false),
		}),
		"option without resolved user": targetInteraction(discordgo.ApplicationCommandInteractionData{
			Options: subCommandWith(memberValue("u2")), Resolved: resolved(true),
		}),
		"no author to fall back on": withoutAuthor,
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := GetTargetMember(event, memberOption)

			require.ErrorIs(t, err, commands.ErrInvalidInteraction)
			assert.ErrorIs(t, err, commands.ErrStaleInteraction)
		})
	}
}
