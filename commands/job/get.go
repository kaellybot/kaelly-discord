package job

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/rs/zerolog/log"
)

func (command *JobCommand) getRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	job, server, err := command.getGetOptions(ctx)
	if err != nil {
		panic(err)
	}

	members, err := s.GuildMembers(i.GuildID, "", memberListLimit)
	if err != nil {
		panic(err)
	}

	var userIds []string
	for _, member := range members {
		userIds = append(userIds, member.User.ID)
	}

	msg := mappers.MapBookJobGetRequest(job.Id, server.Id, userIds, craftsmenListLimit, lg)
	err = command.requestManager.Request(s, i, jobRequestRoutingKey, msg, command.getRespond)
	if err != nil {
		panic(err)
	}
}

func (command *JobCommand) getRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage) {

	if message.Status == amqp.RabbitMQMessage_SUCCESS {

		// TODO respond
		content := fmt.Sprintf("%v", message.JobGetAnswer.Craftsmen)
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
