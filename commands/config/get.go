package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/mappers"
)

func (command *ConfigCommand) getRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapConfigurationGetRequest(i.Interaction.GuildID, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.getRespond)
	if err != nil {
		panic(err)
	}
}

func (command *ConfigCommand) getRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage) {

	// TODO respond
}
