package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/panics"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/rs/zerolog/log"
)

func New(token string, shardID, shardCount int, commands []commands.DiscordCommand,
	broker amqp.MessageBroker, requestManager requests.RequestManager) (*Impl, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error().Err(err).Msgf("Connecting to Discord gateway failed")
		return nil, err
	}

	service := Impl{
		session:        dg,
		commands:       commands,
		broker:         broker,
		requestManager: requestManager,
	}

	dg.Identify.Shard = &[2]int{shardID, shardCount}
	dg.Identify.Intents = constants.GetIntents()
	dg.AddHandler(service.ready)
	dg.AddHandler(service.interactionCreate)
	dg.AddHandler(service.guildCreate)
	dg.AddHandler(service.guildDelete)

	return &service, nil
}

func (service *Impl) Listen() error {
	err := service.session.Open()
	if err != nil {
		log.Error().Int(constants.LogShard, service.session.ShardID).Err(err).Msgf("Impossible to listen events")
		return err
	}

	log.Info().Int(constants.LogShard, service.session.ShardID).Msgf("Discord service is listening events...")
	return nil
}

func (service *Impl) IsConnected() bool {
	return service.session != nil && service.session.DataReady
}

func (service *Impl) Shutdown() {
	log.Info().Int(constants.LogShard, service.session.ShardID).Msgf("Closing Discord connections...")
	err := service.session.Close()
	if err != nil {
		log.Warn().Err(err).Msgf("Cannot close session and shutdown correctly")
	}
}

func (service *Impl) ready(session *discordgo.Session, _ *discordgo.Ready) {
	log.Info().
		Int(constants.LogShard, session.ShardID).
		Int(constants.LogGuildCount, len(session.State.Guilds)).
		Msgf("Ready!")
	err := session.UpdateGameStatus(0, constants.GetGame().Name)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot update the game status, continuing...")
	}
}

func (service *Impl) interactionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	defer panics.HandlePanic(session, event)

	err := service.deferInteraction(session, event)
	if err != nil {
		panic(err)
	}

	for _, command := range service.commands {
		if command.Matches(event) {
			command.Handle(session, event)
			return
		}
	}
}

func (service *Impl) guildCreate(session *discordgo.Session, event *discordgo.GuildCreate) {
	// Ignore outage.
	if event.Unavailable {
		return
	}

	i := discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID: event.ID,
		},
	}

	request := mappers.MapConfigurationGuildCreateRequest(event.Guild, event.MemberCount)
	errReq := service.requestManager.Request(session, &i, constants.ConfigurationRequestRoutingKey,
		request, service.guildCreateReply)
	if errReq != nil {
		log.Error().Err(errReq).
			Str(constants.LogGuildID, event.Guild.ID).
			Msg("Cannot send guild delete event as request, ignoring it...")
	}
}

func (service *Impl) guildCreateReply(_ context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	answer := message.ConfigurationGuildCreateAnswer
	if message.Status == amqp.RabbitMQMessage_SUCCESS && answer != nil && answer.Created {
		newsMessage := mappers.MapGuildCreateNews(answer.GetId(), answer.GetName(), answer.MemberCount)
		errBroker := service.broker.Emit(newsMessage, amqp.ExchangeNews, constants.GuildNewsRoutingKey, i.ID)
		if errBroker != nil {
			log.Warn().Err(errBroker).
				Msgf("Cannot trace guild create through AMQP, continuing...")
		}

		// TODO Send welcome message in the first channel where we have the right to.
		// In case we don't, try to send message to ownerID
		log.Info().Msg("WELCOME")
	}
}

func (service *Impl) guildDelete(session *discordgo.Session, event *discordgo.GuildDelete) {
	// Ignore outage.
	if event.Unavailable {
		return
	}

	i := discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID: event.ID,
		},
	}

	request := mappers.MapConfigurationGuildDeleteRequest(event.BeforeDelete, event.BeforeDelete.MemberCount)
	errReq := service.requestManager.Request(session, &i, constants.ConfigurationRequestRoutingKey,
		request, service.guildDeleteReply)
	if errReq != nil {
		log.Warn().Err(errReq).
			Str(constants.LogGuildID, event.BeforeDelete.ID).
			Msg("Cannot send guild delete event as request, ignoring it...")
	}
}

func (service *Impl) guildDeleteReply(_ context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	answer := message.ConfigurationGuildDeleteAnswer
	if message.Status == amqp.RabbitMQMessage_SUCCESS && answer != nil && answer.Deleted {
		newsMessage := mappers.MapGuildDeleteNews(answer.GetId(), answer.GetName(), answer.MemberCount)
		errBroker := service.broker.Emit(newsMessage, amqp.ExchangeNews, constants.GuildNewsRoutingKey, i.ID)
		if errBroker != nil {
			log.Warn().Err(errBroker).Msgf("Cannot trace guild delete through AMQP, continuing...")
		}
	}
}

func (service *Impl) deferInteraction(session *discordgo.Session,
	event *discordgo.InteractionCreate) error {
	if event.Interaction.Type == discordgo.InteractionApplicationCommand {
		return session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
	} else if event.Interaction.Type == discordgo.InteractionMessageComponent {
		return session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
	}

	return nil
}
