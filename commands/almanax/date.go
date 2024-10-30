//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
package almanax

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/rs/zerolog/log"
)

func (command *Command) getAlmanax(ctx context.Context, s *discordgo.Session,
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

func (command *Command) updateAlmanax(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	date, ok := contract.ExtractAlmanaxDayCustomID(customID)
	if !ok {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	msg := mappers.MapAlmanaxRequest(date, i.Locale)
	err := command.requestManager.Request(s, i, almanaxRequestRoutingKey, msg, command.almanaxRespond)
	if err != nil {
		panic(err)
	}
}

func (command *Command) updateAlmanaxByDate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	seconds, errConv := discord.GetInt64Value(i.MessageComponentData())
	if errConv != nil {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, i.MessageComponentData().CustomID).
			Strs(constants.LogRequestValue, i.MessageComponentData().Values).
			Msgf("Cannot retrieve duration from values selected by user, panicking...")
		panic(errConv)
	}

	day := time.Unix(seconds, 0).UTC()
	msg := mappers.MapAlmanaxRequest(&day, i.Locale)
	err := command.requestManager.Request(s, i, almanaxRequestRoutingKey, msg, command.almanaxRespond)
	if err != nil {
		panic(err)
	}
}

func (command *Command) almanaxRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isAlmanaxAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	webhookEdit := mappers.MapAlmanaxToWebhook(message.GetEncyclopediaAlmanaxAnswer(),
		constants.MapAMQPLocale(message.Language), command.emojiService)
	_, err := s.InteractionResponseEdit(i.Interaction, webhookEdit)
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
