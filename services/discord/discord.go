package discord

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/utils/panics"
	"github.com/rs/zerolog/log"
)

func New(token string, shardID, shardCount int, commands []commands.DiscordCommand,
	guildService guilds.Service, broker amqp.MessageBroker) (*Impl, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error().Err(err).Msgf("Connecting to Discord gateway failed")
		return nil, err
	}

	service := Impl{
		session:      dg,
		commands:     commands,
		guildService: guildService,
		broker:       broker,
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

func (service *Impl) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	// Ignore outage.
	if event.Unavailable {
		return
	}

	guild := event.Guild
	exists, errExist := service.guildService.Exists(guild.ID)
	if errExist != nil {
		log.Error().Err(errExist).
			Str(constants.LogGuildID, guild.ID).
			Msg("Cannot check guild existence, ignoring create event")
		return
	}

	// Ignore already existing guilds
	if exists {
		return
	}

	newsMessage := mappers.MapGuildCreateNews(guild.ID, guild.Name, guild.MemberCount)
	errEmit := service.broker.Emit(newsMessage, amqp.ExchangeNews, constants.GuildNewsRoutingKey, event.ID)
	if errEmit != nil {
		log.Warn().Err(errEmit).
			Msgf("Cannot emit guild create event through AMQP, continuing...")
	}

	service.welcomeGuild(s, guild)
}

func (service *Impl) guildDelete(_ *discordgo.Session, event *discordgo.GuildDelete) {
	// Ignore outage.
	if event.Unavailable {
		return
	}

	guild := event.BeforeDelete
	exists, errExist := service.guildService.Exists(guild.ID)
	if errExist != nil {
		log.Error().Err(errExist).
			Str(constants.LogGuildID, guild.ID).
			Msg("Cannot check guild existence, ignoring delete event")
		return
	}

	// Ignore already deleted guilds
	if !exists {
		return
	}

	newsMessage := mappers.MapGuildDeleteNews(guild.ID, guild.Name, guild.MemberCount)
	errEmit := service.broker.Emit(newsMessage, amqp.ExchangeNews, constants.GuildNewsRoutingKey, event.ID)
	if errEmit != nil {
		log.Warn().Err(errEmit).
			Msgf("Cannot emit guild delete event through AMQP, continuing...")
	}
}
