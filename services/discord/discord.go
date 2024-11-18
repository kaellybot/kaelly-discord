package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/panics"
	"github.com/rs/zerolog/log"
)

func New(token string, shardID, shardCount int, commands []commands.DiscordCommand) (*Impl, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error().Err(err).Msgf("Connecting to Discord gateway failed")
		return nil, err
	}

	service := Impl{
		session:  dg,
		commands: commands,
	}

	dg.Identify.Shard = &[2]int{shardID, shardCount}
	dg.Identify.Intents = constants.GetIntents()
	dg.AddHandler(service.ready)
	dg.AddHandler(service.interactionCreate)

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
