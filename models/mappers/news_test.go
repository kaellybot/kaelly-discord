package mappers

import (
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapGuildNews_CarriesGuildAndEventWithoutLanguage(t *testing.T) {
	cases := []struct {
		event   amqp.NewsGuildMessage_Event
		message *amqp.RabbitMQMessage
	}{
		{amqp.NewsGuildMessage_CREATE, MapGuildCreateNews("g1", "My guild", 42)},
		{amqp.NewsGuildMessage_DELETE, MapGuildDeleteNews("g1", "My guild", 42)},
	}
	for _, tc := range cases {
		event, message := tc.event, tc.message
		t.Run(event.String(), func(t *testing.T) {
			assert.Equal(t, amqp.RabbitMQMessage_NEWS_GUILD, message.Type)
			assert.Equal(t, amqp.Language_ANY, message.Language)
			assert.Equal(t, amqp.Game_DOFUS_GAME, message.Game)
			assert.Empty(t, message.UserID)
			require.NotNil(t, message.NewsGuildMessage)
			assert.Equal(t, "g1", message.NewsGuildMessage.Id)
			assert.Equal(t, "My guild", message.NewsGuildMessage.Name)
			assert.Equal(t, int64(42), message.NewsGuildMessage.MemberCount)
			assert.Equal(t, event, message.NewsGuildMessage.Event)
		})
	}
}
