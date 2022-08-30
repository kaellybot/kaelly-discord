package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	"github.com/rs/zerolog/log"
)

type DiscordService interface {
	Listen() error
	RegisterCommands() error
	Shutdown() error
}

type DiscordServiceImpl struct {
	session *discordgo.Session
}

func New(token string, shardID, shardCount int) (*DiscordServiceImpl, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error().Err(err).Msgf("Connecting to Discord gateway failed")
		return nil, err
	}

	service := DiscordServiceImpl{
		session: dg,
	}

	dg.Identify.Shard = &[2]int{shardID, shardCount}
	dg.Identify.Intents = models.Intents
	dg.AddHandler(service.ready)
	dg.AddHandler(service.messageCreate)

	return &service, nil
}

func (service *DiscordServiceImpl) Listen() error {
	err := service.session.Open()
	if err != nil {
		log.Error().Int(models.LogShard, service.session.ShardID).Err(err).Msgf("Impossible to listen events")
		return err
	}

	log.Info().Int(models.LogShard, service.session.ShardID).Msgf("Discord service is listening events...")
	return nil
}

func (service *DiscordServiceImpl) RegisterCommands() error {
	// TODO
	return nil
}

func (service *DiscordServiceImpl) Shutdown() error {
	log.Info().Int(models.LogShard, service.session.ShardID).Msgf("Closing Discord connections...")
	return service.session.Close()
}

func (service *DiscordServiceImpl) ready(session *discordgo.Session, event *discordgo.Ready) {
	log.Info().Int(models.LogShard, session.ShardID).Int(models.LogGuildCount, len(session.State.Guilds)).Msgf("Ready!")
	session.UpdateGameStatus(0, models.Game)
}

func (service *DiscordServiceImpl) messageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	defer service.handlePanicSilently()

	if !event.Author.Bot {
		// do stuff
		if event.Author.ID == "162842827183751169" {
			session.ChannelMessageSend(event.ChannelID, "coucou")
		}
	}
}

func (service *DiscordServiceImpl) handlePanicSilently() {
	r := recover()
	if r == nil {
		return
	}
	log.Error().Msgf("Panic detected: %v", r)
}

func (service *DiscordServiceImpl) handlePanic(channelID string) {
	r := recover()
	if r == nil {
		return
	}
	log.Error().Msgf("Panic detected: %v", r)
	_, err := service.session.ChannelMessageSend(channelID, fmt.Sprintf("oops, an error occurred: %v", r))
	if err != nil {
		log.Warn().Err(err).Msgf("Could not respond to caller after panicking")
	}
}
