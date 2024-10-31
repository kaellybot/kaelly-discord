package align

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/rs/zerolog/log"
)

func (command *Command) getBook(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	city, order, server, errOpt := getGetOptions(ctx)
	if errOpt != nil {
		panic(errOpt)
	}

	properties, errMembers := discord.GetMemberNickNames(s, i.GuildID)
	if errMembers != nil {
		panic(errMembers)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	msg := mappers.MapBookAlignGetBookRequest(city.ID, order.ID, server.ID,
		constants.DefaultPage, userIDs, i.Locale)
	errReq := command.requestManager.Request(s, i, alignRequestRoutingKey, msg,
		command.getBookReply, properties)
	if errReq != nil {
		panic(errReq)
	}
}

func (command *Command) updateBook(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	cityID, orderID, serverID, page, ok := contract.ExtractAlignBookCustomID(customID)
	if !ok {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	properties, errMembers := discord.GetMemberNickNames(s, i.GuildID)
	if errMembers != nil {
		panic(errMembers)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	msg := mappers.MapBookAlignGetBookRequest(cityID, orderID, serverID,
		page, userIDs, i.Locale)
	errReq := command.requestManager.Request(s, i, alignRequestRoutingKey, msg,
		command.getBookReply, properties)
	if errReq != nil {
		panic(errReq)
	}
}

func (command *Command) getBookReply(_ context.Context, s *discordgo.Session,
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
