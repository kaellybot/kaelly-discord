package almanax

import (
	"context"
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
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	duration, err := getDurationOption(ctx)
	if err != nil {
		panic(err)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxResourceRequest(duration, authorID, i.Locale)
	err = command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.getResourcesReply)
	if err != nil {
		panic(err)
	}
}

func (command *Command) updateResourceCharacter(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	dayDuration, ok := contract.ExtractAlmanaxResourceCharacterCustomID(customID)
	if !ok {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	characterNumber, errConv := discord.GetInt64Value(i.MessageComponentData())
	if errConv != nil {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Str(constants.LogRequestProperty, characterNumberProperty).
			Strs(constants.LogRequestValue, i.MessageComponentData().Values).
			Msgf("Cannot retrieve duration from values selected by user, panicking...")
		panic(errConv)
	}

	properties := map[string]any{
		characterNumberProperty: sanitizeCharacterNumber(characterNumber),
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxResourceRequest(sanitizeDayDuration(dayDuration), authorID, i.Locale)
	errReq := command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.updateResourcesReply, properties)
	if errReq != nil {
		panic(errReq)
	}
}

func (command *Command) updateResourceDuration(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	characterNumber, ok := contract.ExtractAlmanaxResourceDurationCustomID(customID)
	if !ok {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	properties := map[string]any{
		characterNumberProperty: sanitizeCharacterNumber(characterNumber),
	}

	duration, errConv := discord.GetInt64Value(i.MessageComponentData())
	if errConv != nil {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Str(constants.LogRequestProperty, dayDurationProperty).
			Strs(constants.LogRequestValue, i.MessageComponentData().Values).
			Msgf("Cannot retrieve duration from values selected by user, panicking...")
		panic(errConv)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxResourceRequest(sanitizeDayDuration(duration), authorID, i.Locale)
	errReq := command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
		msg, command.updateResourcesReply, properties)
	if errReq != nil {
		panic(errReq)
	}
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
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	command.updateResourcesReply(ctx, s, i, message, map[string]any{
		characterNumberProperty: int64(defaultCharacterNumber),
	})
}

func (command *Command) updateResourcesReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {
	if !isAlmanaxResourceAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	characterNumberValue, found := properties[characterNumberProperty]
	if !found {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogRequestProperty, characterNumberProperty).
			Msgf("Cannot find request property, panicking...")
		panic(commands.ErrRequestPropertyNotFound)
	}
	characterNumber, ok := characterNumberValue.(int64)
	if !ok {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogRequestProperty, characterNumberProperty).
			Msgf("Cannot convert request property, panicking...")
		panic(commands.ErrRequestPropertyNotFound)
	}

	webhookEdit := mappers.MapAlmanaxResourceToWebhook(message.GetEncyclopediaAlmanaxResourceAnswer(),
		characterNumber, i18n.MapAMQPLocale(message.Language), command.emojiService)

	_, err := s.InteractionResponseEdit(i.Interaction, webhookEdit)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func isAlmanaxResourceAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_ANSWER &&
		message.GetEncyclopediaAlmanaxResourceAnswer() != nil
}
