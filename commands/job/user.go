package job

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/rs/zerolog/log"
)

func (command *Command) userJobRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate) error {
	server, err := getUserOptions(ctx)
	if err != nil {
		return err
	}

	member, memberFound := i.ApplicationCommandData().Resolved.Members[i.ApplicationCommandData().TargetID]
	user, userFound := i.ApplicationCommandData().Resolved.Users[i.ApplicationCommandData().TargetID]
	if !memberFound || !userFound {
		return fmt.Errorf("%w: target %q is not resolved",
			commands.ErrInvalidInteraction, i.ApplicationCommandData().TargetID)
	}
	member.User = user

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookJobGetUserRequest(member.User.ID, server.ID, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.JobRequestRoutingKey,
		msg, command.userRespond,
		map[string]any{userProperty: member})
}

func (command *Command) userRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) error {
	if message.Status != amqp.RabbitMQMessage_SUCCESS {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	member, ok := properties[userProperty].(*discordgo.Member)
	if !ok {
		return fmt.Errorf("%w: %v", commands.ErrRequestPropertyNotFound, userProperty)
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: mappers.MapJobUserToEmbed(message.JobGetUserAnswer.Jobs, member, message.JobGetUserAnswer.ServerId,
			command.bookService, command.serverService, message.Language),
	})
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}
