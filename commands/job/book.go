package job

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
	job, server, err := getGetOptions(ctx)
	if err != nil {
		return err
	}

	properties, err := discord.GetMemberNickNames(s, i.GuildID)
	if err != nil {
		// Discord refusing the member list is someone else failing, not our
		// logic: worth a retry message rather than a crash report.
		return fmt.Errorf("%w: %w", commands.ErrUnavailable, err)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookJobGetBookRequest(job.ID, server.ID,
		constants.DefaultPage, userIDs, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.JobRequestRoutingKey,
		msg, command.getBookReply, properties)
}

func (command *Command) updateBookPage(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	jobID, serverID, page, ok := contract.ExtractJobBookPageCustomID(customID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	properties, err := discord.GetMemberNickNames(s, i.GuildID)
	if err != nil {
		// Discord refusing the member list is someone else failing, not our
		// logic: worth a retry message rather than a crash report.
		return fmt.Errorf("%w: %w", commands.ErrUnavailable, err)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookJobGetBookRequest(jobID, serverID,
		page, userIDs, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.JobRequestRoutingKey,
		msg, command.getBookReply, properties)
}

func (command *Command) updateJobBook(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	values := i.MessageComponentData().Values
	if len(values) != 1 {
		return fmt.Errorf("%w: custom ID %q carries %v selected jobs",
			commands.ErrInvalidInteraction, customID, len(values))
	}

	serverID, ok := contract.ExtractJobBookSelectCustomID(customID)
	if !ok {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	properties, err := discord.GetMemberNickNames(s, i.GuildID)
	if err != nil {
		// Discord refusing the member list is someone else failing, not our
		// logic: worth a retry message rather than a crash report.
		return fmt.Errorf("%w: %w", commands.ErrUnavailable, err)
	}

	var userIDs []string
	for userID := range properties {
		userIDs = append(userIDs, userID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapBookJobGetBookRequest(values[0], serverID,
		constants.DefaultPage, userIDs, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.JobRequestRoutingKey,
		msg, command.getBookReply, properties)
}

func (command *Command) getBookReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) error {
	if !isJobGetBookAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	craftsmen := make([]constants.JobUserLevel, 0)
	for _, craftsman := range message.JobGetBookAnswer.Craftsmen {
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

	webhook := mappers.MapJobBookToWebhook(message.GetJobGetBookAnswer(), craftsmen,
		command.bookService, command.serverService, command.emojiService,
		i18n.MapAMQPLocale(message.Language))
	_, err := s.InteractionResponseEdit(i.Interaction, webhook)
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}

func isJobGetBookAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_JOB_GET_BOOK_ANSWER &&
		message.GetJobGetBookAnswer() != nil
}
