package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) serverRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	server, channelID, err := getServerOptions(ctx)
	if err != nil {
		panic(err)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapConfigurationServerRequest(i.GuildID, channelID, server.ID, authorID, i.Locale)
	err = command.requestManager.Request(s, i, constants.ConfigurationRequestRoutingKey,
		msg, command.serverRespond)
	if err != nil {
		panic(err)
	}
}

func (command *Command) serverRespond(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	if !isConfigSetServerAnswerValid(message) {
		panic(commands.ErrInvalidAnswerMessage)
	}

	content := i18n.Get(constants.MapAMQPLocale(message.Language), "config.success")
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}
}

func isConfigSetServerAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Type == amqp.RabbitMQMessage_CONFIGURATION_SET_SERVER_ANSWER &&
		message.Status == amqp.RabbitMQMessage_SUCCESS
}
