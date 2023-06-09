package mappers

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

func MapItemListRequest(query string, lg discordgo.Locale) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_LIST_REQUEST,
		Language: constants.MapDiscordLocale(lg),
		EncyclopediaItemListRequest: &amqp.EncyclopediaItemListRequest{
			Query: query,
		},
	}
}
