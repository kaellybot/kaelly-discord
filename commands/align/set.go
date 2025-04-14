package align

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	di18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) setBook(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	city, order, level, server, err := getSetOptions(ctx)
	if err != nil {
		panic(err)
	}

	userID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookAlignSetRequest(userID, city.ID, order.ID, server.ID, level, i.Locale)
	err = command.requestManager.Request(s, i, constants.AlignRequestRoutingKey,
		msg, command.setBookReply)
	if err != nil {
		panic(err)
	}
}

func (command *Command) setBookReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if message.Status == amqp.RabbitMQMessage_SUCCESS {
		content := di18n.Get(i18n.MapAMQPLocale(message.Language), "align.success")
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}
	} else {
		panic(commands.ErrInvalidAnswerMessage)
	}
}
