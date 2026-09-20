package almanax

import (
	"context"
	"fmt"
	"slices"

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

func (command *Command) getResources(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	duration, err := getDurationOption(ctx)
	if err != nil {
		return err
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxResourceRequest(duration, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.getResourcesReply)
}

func (command *Command) updateResourceCharacter(s *discordgo.Session,
	i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()
	dayDuration, ok := contract.ExtractAlmanaxResourceCharacterCustomID(data.CustomID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, data.CustomID)
	}

	characterNumber, errConv := discord.GetInt64Value(data)
	if errConv != nil {
		return fmt.Errorf("cannot read the %v selected on custom ID %q among %v: %w",
			characterNumberProperty, data.CustomID, data.Values, errConv)
	}

	properties := map[string]any{
		characterNumberProperty: sanitizeCharacterNumber(characterNumber),
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxResourceRequest(sanitizeDayDuration(dayDuration), authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.updateResourcesReply, properties)
}

func (command *Command) updateResourceDuration(s *discordgo.Session,
	i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()
	characterNumber, ok := contract.ExtractAlmanaxResourceDurationCustomID(data.CustomID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, data.CustomID)
	}

	properties := map[string]any{
		characterNumberProperty: sanitizeCharacterNumber(characterNumber),
	}

	duration, errConv := discord.GetInt64Value(data)
	if errConv != nil {
		return fmt.Errorf("cannot read the %v selected on custom ID %q among %v: %w",
			dayDurationProperty, data.CustomID, data.Values, errConv)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxResourceRequest(sanitizeDayDuration(duration), authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.updateResourcesReply, properties)
}

// sanitizeCharacterNumber keeps the character count among the values the menu
// offers. Both the custom ID carrying it and the selected values come from the
// caller, so neither can be trusted; the library stopped clamping it in v1.0.3
// because those values are owned here.
func sanitizeCharacterNumber(characterNumber int64) int64 {
	if slices.Contains(constants.GetCharacterNumbers(), characterNumber) {
		return characterNumber
	}

	return defaultCharacterNumber
}

// sanitizeDayDuration keeps the duration within the bounds the slash command
// advertises; the menu only offers a subset of them, and the command itself
// accepts any value in between.
func sanitizeDayDuration(dayDuration int64) int64 {
	return min(max(dayDuration, contract.AlmanaxDurationMinimumValue), contract.AlmanaxDurationMaximumValue)
}

func (command *Command) getResourcesReply(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) error {
	return command.updateResourcesReply(ctx, s, i, message, map[string]any{
		characterNumberProperty: int64(defaultCharacterNumber),
	})
}

func (command *Command) updateResourcesReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) error {
	if !isAlmanaxResourceAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	characterNumber, ok := properties[characterNumberProperty].(int64)
	if !ok {
		return fmt.Errorf("%w: %v", commands.ErrRequestPropertyNotFound, characterNumberProperty)
	}

	webhookEdit := mappers.MapAlmanaxResourceToWebhook(message.GetEncyclopediaAlmanaxResourceAnswer(),
		characterNumber, i18n.MapAMQPLocale(message.Language), command.emojiService)

	_, err := s.InteractionResponseEdit(i.Interaction, webhookEdit)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}

func isAlmanaxResourceAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_ANSWER &&
		message.GetEncyclopediaAlmanaxResourceAnswer() != nil
}
