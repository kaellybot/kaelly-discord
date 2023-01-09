package job

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/mappers"
)

func (command *JobCommand) setRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	job, level, server, err := command.getSetOptions(ctx)
	if err != nil {
		panic(err)
	}

	msg := mappers.MapBookJobSetRequest(i.Interaction.User.ID, job.Id, server.Id, level, lg)
	err = command.requestManager.Request(s, i, jobRequestRoutingKey, msg, command.setRespond)
	if err != nil {
		panic(err)
	}
}

func (command *JobCommand) setRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage) {

	// TODO respond
}
