package job

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
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

	msg := mappers.MapBookJobSetRequest(i.Interaction.Member.User.ID, job.Id, server.Id, level, lg)
	err = command.requestManager.Request(s, i, jobRequestRoutingKey, msg, command.setRespond)
	if err != nil {
		panic(err)
	}
}

func (command *JobCommand) setRespond(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) {

	if message.Status == amqp.RabbitMQMessage_SUCCESS {
		content := i18n.Get(constants.MapAmqpLocale(message.Language), "job.success")
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
		}
	} else {
		panic(commands.ErrInvalidAnswerMessage)
	}
}
