package align

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *AlignCommand) setRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	city, order, level, server, err := command.getSetOptions(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapBookAlignSetRequest(i.Interaction.Member.User.ID, city.Id, order.Id, server.Id, level, lg)
	err = command.requestManager.Request(s, i, alignRequestRoutingKey, msg, command.setRespond)
	if err != nil {
		panic(err)
	}
}

func (command *AlignCommand) setRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	if message.Status == amqp.RabbitMQMessage_SUCCESS {
		content := i18n.Get(constants.MapAmqpLocale(message.Language), "align.success")
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
