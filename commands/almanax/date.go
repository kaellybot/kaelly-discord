//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
package almanax

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) almanaxRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	date, err := getDateOption(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapAlmanaxRequest(date, i.Locale)
	err = command.requestManager.Request(s, i, almanaxRequestRoutingKey, msg, command.almanaxRespond)
	if err != nil {
		panic(err)
	}
}

func (command *Command) almanaxRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isAlmanaxAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	lg := constants.MapAMQPLocale(message.Language)
	webhookEdit := discordgo.WebhookEdit{}
	if message.GetEncyclopediaAlmanaxAnswer().Almanax != nil {
		webhookEdit.Embeds = &[]*discordgo.MessageEmbed{
			mappers.MapAlmanaxToEmbed(message.GetEncyclopediaAlmanaxAnswer().Almanax, lg),
		}
	} else {
		content := i18n.Get(lg, "almanax.day.missing")
		webhookEdit.Content = &content
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &webhookEdit)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func isAlmanaxAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_ANSWER &&
		message.GetEncyclopediaAlmanaxAnswer() != nil
}
