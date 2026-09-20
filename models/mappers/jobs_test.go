package mappers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jobBooks() fakeBooks {
	return fakeBooks{jobs: []entities.Job{
		{ID: "baker", Icon: "baker.png", Color: 333, Labels: []entities.JobLabel{
			{JobID: "baker", Locale: amqp.Language_EN, Label: "Baker"},
			{JobID: "baker", Locale: amqp.Language_FR, Label: "Boulanger"},
		}},
		{ID: "alchemist", Icon: "alchemist.png", Color: 444, Labels: []entities.JobLabel{
			{JobID: "alchemist", Locale: amqp.Language_EN, Label: "Alchemist"},
			{JobID: "alchemist", Locale: amqp.Language_FR, Label: "Alchimiste"},
		}},
	}}
}

func TestMapBookJobRequests_CopyFieldsAndComputeOffset(t *testing.T) {
	book := MapBookJobGetBookRequest("baker", "srv", 3, []string{"u1"}, "author", discordgo.French)
	assert.Equal(t, amqp.RabbitMQMessage_JOB_GET_BOOK_REQUEST, book.Type)
	assert.Equal(t, amqp.Language_FR, book.Language)
	assert.Equal(t, "author", book.UserID)
	require.NotNil(t, book.JobGetBookRequest)
	assert.Equal(t, []string{"u1"}, book.JobGetBookRequest.UserIds)
	assert.Equal(t, "baker", book.JobGetBookRequest.JobId)
	assert.Equal(t, "srv", book.JobGetBookRequest.ServerId)
	assert.Equal(t, int64(3*constants.MaxBookRowPerEmbed), book.JobGetBookRequest.Offset)
	assert.Equal(t, int64(constants.MaxBookRowPerEmbed), book.JobGetBookRequest.Size)

	user := MapBookJobGetUserRequest("target", "srv", "author", discordgo.SpanishES)
	assert.Equal(t, amqp.RabbitMQMessage_JOB_GET_USER_REQUEST, user.Type)
	assert.Equal(t, amqp.Language_ES, user.Language)
	assert.Equal(t, "author", user.UserID)
	require.NotNil(t, user.JobGetUserRequest)
	assert.Equal(t, "target", user.JobGetUserRequest.UserId)
	assert.Equal(t, "srv", user.JobGetUserRequest.ServerId)

	set := MapBookJobSetRequest("user", "baker", "srv", 200, discordgo.PortugueseBR)
	assert.Equal(t, amqp.RabbitMQMessage_JOB_SET_REQUEST, set.Type)
	assert.Equal(t, amqp.Language_PT, set.Language)
	assert.Equal(t, "user", set.UserID, "the user sets their own level")
	require.NotNil(t, set.JobSetRequest)
	assert.Equal(t, "user", set.JobSetRequest.UserId)
	assert.Equal(t, "baker", set.JobSetRequest.JobId)
	assert.Equal(t, "srv", set.JobSetRequest.ServerId)
	assert.Equal(t, int64(200), set.JobSetRequest.Level)
}

func TestMapJobBookToWebhook_ListsCraftsmenByLevelThenName(t *testing.T) {
	answer := &amqp.JobGetBookAnswer{JobId: "baker", ServerId: "srv", Page: 0, Pages: 2, Total: 20}
	craftsmen := []constants.JobUserLevel{
		{Username: "Zed", Level: 100},
		{Username: "Max", Level: 200},
		{Username: "Abe", Level: 100},
	}

	edit := MapJobBookToWebhook(answer, craftsmen, jobBooks(),
		fakeServers{servers: []entities.Server{testServer()}}, fakeEmojis{}, discordgo.French)

	embed := singleEmbed(t, edit.Embeds)
	assert.Contains(t, embed.Title, "Boulanger")
	assertRendered(t, embed.Title)
	assertRendered(t, embed.Description)
	assert.Contains(t, embed.Description, "20")
	assertInOrder(t, embed.Description, "**Max**", "**Abe**", "**Zed**")
	assert.Equal(t, 333, embed.Color)
	assert.Equal(t, "baker.png", embed.Thumbnail.URL)
	assert.Equal(t, "Imagiro-FR", embed.Footer.Text)
	assert.Equal(t, "srv.png", embed.Footer.IconURL)

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 2)
	pagination := buttons(t, rows[0])
	require.Len(t, pagination, 2)
	assert.True(t, pagination[0].Disabled, "first page cannot go back")
	jobID, serverID, page, ok := contract.ExtractJobBookPageCustomID(pagination[1].CustomID)
	require.True(t, ok, pagination[1].CustomID)
	assert.Equal(t, []any{"baker", "srv", 1}, []any{jobID, serverID, page})

	menu := selectMenu(t, rows[1])
	menuServerID, ok := contract.ExtractJobBookSelectCustomID(menu.CustomID)
	require.True(t, ok, menu.CustomID)
	assert.Equal(t, "srv", menuServerID)
	assertTranslated(t, discordgo.French, "job.embed.choices.placeholder", menu.Placeholder)
	assert.Equal(t, []discordgo.SelectMenuOption{
		{Label: "Boulanger", Value: "baker", Default: true, Emoji: &discordgo.ComponentEmoji{Name: "Job:baker"}},
		{Label: "Alchimiste", Value: "alchemist", Emoji: &discordgo.ComponentEmoji{Name: "Job:alchemist"}},
	}, menu.Options)
}

func TestMapJobBookToWebhook_UnknownJobAndServerAndNoCraftsmen(t *testing.T) {
	answer := &amqp.JobGetBookAnswer{JobId: "ghost-job", ServerId: "ghost-srv", Pages: 1}

	edit := MapJobBookToWebhook(answer, nil, jobBooks(), fakeServers{}, fakeEmojis{}, discordgo.EnglishGB)

	embed := singleEmbed(t, edit.Embeds)
	assert.Contains(t, embed.Title, "ghost-job", "unknown job falls back to its ID")
	assertRendered(t, embed.Description)
	assert.NotContains(t, embed.Description, "**")
	assert.Zero(t, embed.Color)
	assert.Equal(t, "ghost-srv", embed.Footer.Text)

	rows := actionRows(t, edit.Components)
	require.Len(t, rows, 1, "a single page has no pagination row")
	for _, option := range selectMenu(t, rows[0]).Options {
		assert.False(t, option.Default)
	}
}

func TestMapJobUserToEmbed_SortsJobsByLevelThenLocalizedLabel(t *testing.T) {
	member := &discordgo.Member{Nick: "Nick", User: &discordgo.User{ID: "u", Username: "user", Avatar: "useravatar"}}
	experiences := []*amqp.JobGetUserAnswer_JobExperience{
		{JobId: "ghost", Level: 1},
		{JobId: "baker", Level: 50},
		{JobId: "alchemist", Level: 50},
	}

	embeds := MapJobUserToEmbed(experiences, member, "srv", jobBooks(),
		fakeServers{servers: []entities.Server{testServer()}}, amqp.Language_FR)

	embed := singleEmbed(t, embeds)
	assert.Contains(t, embed.Title, "Nick")
	assertRendered(t, embed.Description)
	assertInOrder(t, embed.Description, "Alchimiste", "Boulanger", "ghost")
	assert.Equal(t, constants.Color, embed.Color)
	assert.Contains(t, embed.Thumbnail.URL, "useravatar", "member without guild avatar uses the user avatar")
	assert.Equal(t, "Imagiro-FR", embed.Footer.Text)
}

func TestMapJobUserToEmbed_NoJobsAndUnknownServer(t *testing.T) {
	member := &discordgo.Member{User: &discordgo.User{ID: "u", Username: "user"}}

	embeds := MapJobUserToEmbed(nil, member, "ghost", jobBooks(), fakeServers{}, amqp.Language_EN)

	embed := singleEmbed(t, embeds)
	assert.Contains(t, embed.Title, "user")
	assertRendered(t, embed.Description)
	assert.Equal(t, "ghost", embed.Footer.Text)
	assert.Empty(t, embed.Footer.IconURL)
}
