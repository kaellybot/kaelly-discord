package job

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
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
	properties := make(map[string]any)
	for _, member := range members {
		member.Mention()
		userIds = append(userIds, member.User.ID)
		username := member.Nick
		if len(username) == 0 {
			username = member.User.Username
		}
		properties[member.User.ID] = username
	}

	msg := mappers.MapBookJobGetRequest(job.Id, server.Id, userIds, craftsmenListLimit, lg)
	err = command.requestManager.Request(s, i, jobRequestRoutingKey, msg, command.getRespond, properties)
	if err != nil {
		panic(err)
	}
}

func (command *JobCommand) getRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	if message.Status == amqp.RabbitMQMessage_SUCCESS {

		craftsmen := make([]constants.JobUserLevel, 0)
		for _, craftsman := range message.JobGetAnswer.Craftsmen {
			username, found := properties[craftsman.UserId]
			if found {
				craftsmen = append(craftsmen, constants.JobUserLevel{
					Username: fmt.Sprintf("%v", username),
					Level:    craftsman.Level,
				})
			} else {
				log.Warn().Msgf("MemberId not found in property, item ignored...")
			}
		}

		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: mappers.MapJobBookToEmbed(craftsmen, message.JobGetAnswer.JobId, message.JobGetAnswer.ServerId,
				command.bookService, command.serverService, message.Language),
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}
	} else {
		panic(commands.ErrInvalidAnswerMessage)
	}
}
