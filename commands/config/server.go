package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/mappers"
)

func (command *ConfigCommand) serverRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	server, channelId, err := command.getServerOptions(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapConfigurationServerRequest(i.Interaction.GuildID, channelId, server.Id, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.serverRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) serverRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage) {

	// TODO respond
}
