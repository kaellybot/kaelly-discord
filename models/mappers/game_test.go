package mappers

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/stretchr/testify/assert"
)

// Every message this bot publishes must carry its game: the broker refuses to
// publish one without, so a mapper that forgets it breaks its command outright.
// This sweeps all of them at once, so a new mapper is not silently missed.
func TestEveryPublishedMessageCarriesTheGame(t *testing.T) {
	date := time.Now()
	query := "query"

	messages := map[string]*amqp.RabbitMQMessage{
		"about":        MapAboutRequest("author", discordgo.French),
		"help":         MapHelpRequest("author", discordgo.French),
		"almanax":      MapAlmanaxRequest(&date, "author", discordgo.French),
		"almanaxRes":   MapAlmanaxResourceRequest(7, "author", discordgo.French),
		"almanaxFxLst": MapAlmanaxEffectListRequest("q", "author", discordgo.French),
		"almanaxFx":    MapAlmanaxEffectRequest(&query, &date, 0, "author", discordgo.French),
		"itemList":     MapItemListRequest("q", "author", discordgo.French),
		"item":         MapItemRequest("q", false, amqp.ItemType_EQUIPMENT_TYPE, "author", discordgo.French),
		"setList":      MapSetListRequest("q", "author", discordgo.French),
		"alignBook":    MapBookAlignGetBookRequest("city", "order", "server", 0, nil, "author", discordgo.French),
		"alignUser":    MapBookAlignGetUserRequest("user", "server", "author", discordgo.French),
		"alignSet":     MapBookAlignSetRequest("user", "city", "order", "server", 100, discordgo.French),
		"jobBook":      MapBookJobGetBookRequest("job", "server", 0, nil, "author", discordgo.French),
		"jobUser":      MapBookJobGetUserRequest("user", "server", "author", discordgo.French),
		"jobSet":       MapBookJobSetRequest("user", "job", "server", 100, discordgo.French),
		"configGet":    MapConfigurationGetRequest("guild", "author", discordgo.French),
		"guildCreate":  MapGuildCreateNews("guild", "name", 10),
		"guildDelete":  MapGuildDeleteNews("guild", "name", 10),
	}

	for name, message := range messages {
		t.Run(name, func(t *testing.T) {
			assert.True(t, message.IsGameSet(),
				"the broker refuses to publish a message whose game is not set")
		})
	}
}
