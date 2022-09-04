package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	i18n "github.com/kaysoro/discordgo-i18n"
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
	dg.AddHandler(service.interactionCreate)

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
	for _, command := range discordCommands {
		_, err := service.session.ApplicationCommandCreate(service.session.State.User.ID, "299167247279194112", &command.Identity)
		if err != nil {
			log.Error().Err(err).Str(models.LogCommand, command.Identity.Name).Msgf("Failed to create command, registration stopped")
			return err
		}
		log.Info().Str(models.LogCommand, command.Identity.Name).Msgf("Successfully registered!")
	}

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
	// TODO defer service.handlePanic()
}

func (service *DiscordServiceImpl) interactionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	defer service.handlePanic(session, event)

	// TODO not always ApplicationCommandData
	if command, ok := discordCommands[event.ApplicationCommandData().Name]; ok && event != nil {
		command.Handler(session, event)
	}
}

func (service *DiscordServiceImpl) handlePanic(session *discordgo.Session, event *discordgo.InteractionCreate) {
	r := recover()
	if r == nil {
		return
	}

	// TODO not always ApplicationCommandData
	log.Error().Str(models.LogCommand, event.ApplicationCommandData().Name).Interface(models.LogPanic, r).Msgf("")
	err := session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: i18n.Get(event.Locale, "panic"),
		},
	})
	if err != nil {
		log.Warn().Err(err).Msgf("Could not respond to caller after panicking")
	}
}
