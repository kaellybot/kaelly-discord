package mappers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapAboutToWebhook_BuildsSingleEmbedWithLinksAndEmojis(t *testing.T) {
	edit := MapAboutToWebhook(discordgo.EnglishGB, fakeEmojis{})

	assert.Nil(t, edit.Content)
	assert.Nil(t, edit.Components)
	embed := singleEmbed(t, edit.Embeds)
	assert.Contains(t, embed.Title, constants.Name)
	assert.Contains(t, embed.Title, constants.Version)
	assert.Contains(t, embed.Description, constants.GetGame().Name)
	assert.Contains(t, embed.Description, "<misc:dofusGame>")
	assert.Equal(t, constants.Color, embed.Color)
	assert.Equal(t, constants.AvatarImage, embed.Image.URL)
	assert.Equal(t, constants.GetGame().Icon, embed.Thumbnail.URL)
	require.NotNil(t, embed.Footer)
	assertTranslated(t, discordgo.EnglishGB, "about.footer", embed.Footer.Text)
	assert.Equal(t, constants.AnkamaLogo, embed.Footer.IconURL)

	require.Len(t, embed.Fields, 7)
	for _, field := range embed.Fields {
		assert.False(t, field.Inline)
		assertRendered(t, field.Name)
		assertRendered(t, field.Value)
	}
	support, invite, twitter, github, free, privacy, graphist := embed.Fields[0], embed.Fields[1],
		embed.Fields[2], embed.Fields[3], embed.Fields[4], embed.Fields[5], embed.Fields[6]
	assertTranslated(t, discordgo.EnglishGB, "about.support.title", support.Name)
	assert.Contains(t, support.Value, constants.Discord)
	assertTranslated(t, discordgo.EnglishGB, "about.invite.title", invite.Name)
	for _, link := range []string{constants.InviteDofus, constants.InviteTouch, constants.InviteRetro} {
		assert.Contains(t, invite.Value, "("+link+")")
	}
	assert.Contains(t, twitter.Name, "<misc:twitter>")
	assert.Contains(t, twitter.Value, constants.Twitter)
	assert.Contains(t, github.Name, "<misc:github>")
	assert.Contains(t, github.Value, constants.Github)
	assertTranslated(t, discordgo.EnglishGB, "about.free.title", free.Name)
	assert.Contains(t, free.Value, constants.Paypal)
	assertTranslated(t, discordgo.EnglishGB, "about.privacy.title", privacy.Name)
	assertTranslated(t, discordgo.EnglishGB, "about.privacy.desc", privacy.Value)
	assertTranslated(t, discordgo.EnglishGB, "about.graphist.title", graphist.Name)
	for _, artist := range []constants.Graphist{constants.GetGraphistElycann(), constants.GetGraphistColibry()} {
		assert.Contains(t, graphist.Value, "["+artist.Name+"]("+artist.URL+")")
	}
}

func TestMapAboutToWebhook_UsesRequestedLocale(t *testing.T) {
	edit := MapAboutToWebhook(discordgo.French, fakeEmojis{})

	embed := singleEmbed(t, edit.Embeds)
	assertTranslated(t, discordgo.French, "about.footer", embed.Footer.Text)
	assertTranslated(t, discordgo.French, "about.privacy.title", embed.Fields[5].Name)
}
