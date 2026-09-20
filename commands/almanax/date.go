//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
package almanax

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/i18n"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/rs/zerolog/log"
)

func (command *Command) getAlmanax(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	date, err := getDateOption(ctx)
	if err != nil {
		return err
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxRequest(date, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.almanaxRespond)
}

func (command *Command) updateAlmanax(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	date, ok := contract.ExtractAlmanaxDayCustomID(customID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxRequest(date, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.almanaxRespond)
}

func (command *Command) updateAlmanaxByDate(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()
	seconds, errConv := discord.GetInt64Value(data)
	if errConv != nil {
		return fmt.Errorf("cannot read the day selected on custom ID %q among %v: %w",
			data.CustomID, data.Values, errConv)
	}

	day := time.Unix(seconds, 0).UTC()
	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxRequest(&day, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.almanaxRespond)
}

func (command *Command) almanaxRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) error {
	if !isAlmanaxAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	webhookEdit := mappers.MapAlmanaxToWebhook(message.GetEncyclopediaAlmanaxAnswer(),
		i18n.MapAMQPLocale(message.Language), command.emojiService)
	_, err := s.InteractionResponseEdit(i.Interaction, webhookEdit)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}

func isAlmanaxAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_ANSWER &&
		message.GetEncyclopediaAlmanaxAnswer() != nil
}
