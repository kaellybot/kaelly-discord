package mappers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMapWelcomeToEmbed_MentionsOwnerGuildAndHelpCommand(t *testing.T) {
	previous := viper.Get(constants.HelpID)
	viper.Set(constants.HelpID, "987654")
	t.Cleanup(func() { viper.Set(constants.HelpID, previous) })

	embed := MapWelcomeToEmbed("Les Fous", "owner-1", discordgo.EnglishGB, fakeEmojis{})

	assertRendered(t, embed.Description)
	assert.Contains(t, embed.Description, "</help:987654>", "help is a clickable slash command mention")
	assert.Contains(t, embed.Description, "<@owner-1>")
	assert.Contains(t, embed.Description, "**Les Fous**")
	assert.Contains(t, embed.Description, "`/config`")
	assert.Contains(t, embed.Description, "**"+constants.Name+"**")
	assert.Contains(t, embed.Description, "<misc:dofusGame>")
	assert.Contains(t, embed.Description, "<misc:calendar>")
	assert.Equal(t, constants.Color, embed.Color)
	assert.Equal(t, constants.GetGame().Icon, embed.Thumbnail.URL)
	assert.Equal(t, constants.AvatarImage, embed.Image.URL)
	assertDefaultFooter(t, embed.Footer)
}

func TestMapWelcomeToEmbed_IsLocalized(t *testing.T) {
	english := MapWelcomeToEmbed("G", "o", discordgo.EnglishGB, fakeEmojis{})
	french := MapWelcomeToEmbed("G", "o", discordgo.French, fakeEmojis{})

	assertRendered(t, french.Description)
	assert.NotEqual(t, english.Description, french.Description)
	assert.NotEqual(t, english.Footer.Text, french.Footer.Text)
}
