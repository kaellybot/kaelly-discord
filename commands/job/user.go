package job

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/rs/zerolog/log"
)

func (command *Command) userJobRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {
	server, err := command.getUserOptions(ctx)
	if err != nil {
		panic(err)
	}

	member, memberFound := i.ApplicationCommandData().Resolved.Members[i.ApplicationCommandData().TargetID]
	user, userFound := i.ApplicationCommandData().Resolved.Users[i.ApplicationCommandData().TargetID]
	if !(memberFound && userFound) {
		panic("Cannot retrieve member and user from interaction, panicking")
	}
	member.User = user

	msg := mappers.MapBookJobGetUserRequest(member.User.ID, server.ID, lg)
	err = command.requestManager.Request(s, i, jobRequestRoutingKey, msg, command.userRespond,
		map[string]any{userProperty: member})
	if err != nil {
		panic(err)
	}
}

func (command *Command) userRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {
	if message.Status == amqp.RabbitMQMessage_SUCCESS {
		var member *discordgo.Member
		userProperty, found := properties[userProperty]
		if cast, ok := userProperty.(*discordgo.Member); found && ok {
			member = cast
		} else {
			panic("Member cannot be retrieved from requestHandler properties, panicking")
		}

		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: mappers.MapJobUserToEmbed(message.JobGetUserAnswer.Jobs, member, message.JobGetUserAnswer.ServerId,
				command.bookService, command.serverService, message.Language),
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}
	} else {
		panic(commands.ErrInvalidAnswerMessage)
	}
}
