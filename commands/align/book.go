package align

import (
	"context"
	"fmt"

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

func (command *Command) getBook(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	city, order, server, errOpt := getGetOptions(ctx)
	if errOpt != nil {
		return errOpt
	}

	properties, errMembers := discord.GetMemberNickNames(s, i.GuildID)
	if errMembers != nil {
		return fmt.Errorf("%w: %w", commands.ErrUnavailable, errMembers)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookAlignGetBookRequest(city.ID, order.ID, server.ID,
		constants.DefaultPage, userIDs, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlignRequestRoutingKey,
		msg, command.getBookReply, properties)
}

func (command *Command) updateBookPage(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	cityID, orderID, serverID, page, ok := contract.ExtractAlignBookPageCustomID(customID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	properties, errMembers := discord.GetMemberNickNames(s, i.GuildID)
	if errMembers != nil {
		return fmt.Errorf("%w: %w", commands.ErrUnavailable, errMembers)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookAlignGetBookRequest(cityID, orderID, serverID,
		page, userIDs, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlignRequestRoutingKey,
		msg, command.getBookReply, properties)
}

// nolint:dupl,nolintlint // Simpler to handles these separately.
func (command *Command) updateCityBook(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	values := i.MessageComponentData().Values
	if len(values) != 1 {
		return fmt.Errorf("%w: expected one city value, got %v for custom ID %q",
			commands.ErrInvalidInteraction, len(values), customID)
	}

	var cityID string
	if values[0] != contract.AlignAllValues {
		cityID = values[0]
	}

	orderID, serverID, ok := contract.ExtractAlignBookCityCustomID(customID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	properties, errMembers := discord.GetMemberNickNames(s, i.GuildID)
	if errMembers != nil {
		return fmt.Errorf("%w: %w", commands.ErrUnavailable, errMembers)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookAlignGetBookRequest(cityID, orderID, serverID,
		constants.DefaultPage, userIDs, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlignRequestRoutingKey,
		msg, command.getBookReply, properties)
}

// nolint:dupl,nolintlint // Simpler to handles these separately.
func (command *Command) updateOrderBook(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	values := i.MessageComponentData().Values
	if len(values) != 1 {
		return fmt.Errorf("%w: expected one order value, got %v for custom ID %q",
			commands.ErrInvalidInteraction, len(values), customID)
	}

	var orderID string
	if values[0] != contract.AlignAllValues {
		orderID = values[0]
	}

	cityID, serverID, ok := contract.ExtractAlignBookOrderCustomID(customID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	properties, errMembers := discord.GetMemberNickNames(s, i.GuildID)
	if errMembers != nil {
		return fmt.Errorf("%w: %w", commands.ErrUnavailable, errMembers)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookAlignGetBookRequest(cityID, orderID, serverID,
		constants.DefaultPage, userIDs, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.AlignRequestRoutingKey,
		msg, command.getBookReply, properties)
}

func (command *Command) getBookReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) error {
	if !isAlignGetBookAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
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
		i18n.MapAMQPLocale(message.Language))
	_, err := s.InteractionResponseEdit(i.Interaction, webhook)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}

func isAlignGetBookAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ALIGN_GET_BOOK_ANSWER &&
		message.GetAlignGetBookAnswer() != nil
}
