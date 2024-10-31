package align

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/rs/zerolog/log"
)

func (command *Command) getRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	city, order, server, err := getGetOptions(ctx)
	if err != nil {
		panic(err)
	}

	members, err := s.GuildMembers(i.GuildID, "", memberListLimit)
	if err != nil {
		panic(err)
	}

	var userIDs []string
	properties := make(map[string]any)
	for _, member := range members {
		userIDs = append(userIDs, member.User.ID)
		username := member.Nick
		if len(username) == 0 {
			username = member.User.Username
		}
		properties[member.User.ID] = username
	}

	msg := mappers.MapBookAlignGetBookRequest(city.ID, order.ID, server.ID,
		constants.DefaultPage, userIDs, i.Locale)
	err = command.requestManager.Request(s, i, alignRequestRoutingKey, msg,
		command.getRespond, properties)
	if err != nil {
		panic(err)
	}
}

func (command *Command) getRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {
	if !isAlignGetBookAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	believers := make([]constants.AlignmentUserLevel, 0)
	for _, believer := range message.AlignGetBookAnswer.Believers {
		username, found := properties[believer.UserId]
		if found {
			believers = append(believers, constants.AlignmentUserLevel{
				CityID:   believer.CityId,
				OrderID:  believer.OrderId,
				Username: fmt.Sprintf("%v", username),
				Level:    believer.Level,
			})
		} else {
			log.Warn().Msgf("MemberID not found in property, item ignored...")
		}
	}

	webhook := mappers.MapAlignBookToWebhook(message.GetAlignGetBookAnswer(), believers,
		command.bookService, command.serverService, command.emojiService,
		constants.MapAMQPLocale(message.Language))
	_, err := s.InteractionResponseEdit(i.Interaction, webhook)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func isAlignGetBookAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ALIGN_GET_BOOK_ANSWER &&
		message.GetAlignGetBookAnswer() != nil
}
