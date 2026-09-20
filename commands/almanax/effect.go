//nolint:dupl,nolintlint // OK for DRY concept but refactor at any cost is not relevant here.
package almanax

import (
	"context"
	"fmt"

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

func (command *Command) getAlmanaxesByEffect(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	query, err := getQueryOption(ctx)
	if err != nil {
		return err
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxEffectRequest(&query, nil, constants.DefaultPage, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.effectRespond)
}

func (command *Command) updateAlmanaxesByEffect(s *discordgo.Session,
	i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	day, page, ok := contract.ExtractAlmanaxEffectCustomID(customID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxEffectRequest(nil, day, page, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.effectRespond)
}

func (command *Command) effectRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) error {
	if !isAlmanaxEffectAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	webhookEdit := mappers.MapAlmanaxEffectsToWebhook(message.GetEncyclopediaAlmanaxEffectAnswer(),
		i18n.MapAMQPLocale(message.Language), command.emojiService)
	_, err := s.InteractionResponseEdit(i.Interaction, webhookEdit)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}

func isAlmanaxEffectAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_ANSWER &&
		message.GetEncyclopediaAlmanaxEffectAnswer() != nil
}
