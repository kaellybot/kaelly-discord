package almanax

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/rs/zerolog/log"
)

func (command *Command) getResources(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	duration, err := getDurationOption(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapAlmanaxResourceRequest(duration, i.Locale)
	err = command.requestManager.Request(s, i, almanaxRequestRoutingKey, msg, command.getResourcesReply)
	if err != nil {
		panic(err)
	}
}

func (command *Command) updateResourceCharacter(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO
}

func (command *Command) updateResourceDuration(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO
	/**
	customID := i.MessageComponentData().CustomID
	properties := make(map[string]any)
	duration, characterNumber, ok := contract.ExtractAlmanaxResourceCustomID(customID)
	if !ok {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	msg := mappers.MapCompetitionMapRequest(mapNumber, i.Locale)
	err := command.requestManager.Request(s, i, competitionRequestRoutingKey,
		msg, command.updateMapReply, properties)
	if err != nil {
		panic(err)
	}

	duration, err := getDurationOption(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapAlmanaxResourceRequest(duration, i.Locale)
	err = command.requestManager.Request(s, i, almanaxRequestRoutingKey, msg, command.updateResourcesReply)
	if err != nil {
		panic(err)
	}
	**/
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
		characterNumber, constants.MapAMQPLocale(message.Language), command.emojiService)

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
