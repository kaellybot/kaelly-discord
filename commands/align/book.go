package align

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

func (command *AlignCommand) getRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	city, order, server, err := command.getGetOptions(ctx)
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

	msg := mappers.MapBookAlignGetBookRequest(city.Id, order.Id, server.Id, userIds, believerListLimit, lg)
	err = command.requestManager.Request(s, i, alignRequestRoutingKey, msg, command.getRespond, properties)
	if err != nil {
		panic(err)
	}
}

func (command *AlignCommand) getRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	if message.Status == amqp.RabbitMQMessage_SUCCESS {

		believers := make([]constants.AlignmentUserLevel, 0)
		for _, believer := range message.AlignGetBookAnswer.Believers {
			username, found := properties[believer.UserId]
			if found {
				believers = append(believers, constants.AlignmentUserLevel{
					Username: fmt.Sprintf("%v", username),
					Level:    believer.Level,
				})
			} else {
				log.Warn().Msgf("MemberId not found in property, item ignored...")
			}
		}

		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: mappers.MapAlignBookToEmbed(believers, message.AlignGetBookAnswer.ServerId,
				command.bookService, command.serverService, message.Language),
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}
	} else {
		panic(commands.ErrInvalidAnswerMessage)
	}
}
