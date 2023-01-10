package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/mappers"
)

func (command *ConfigCommand) twitterRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	channelId, enabled, err := command.getWebhookOptions(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapConfigurationWebhookRequest(i.Interaction.GuildID, channelId,
		enabled, amqp.ConfigurationSetRequest_WebhookField_TWITTER, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.twitterRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) twitterRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage) {

	// TODO respond
}
