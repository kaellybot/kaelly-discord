package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type DiscordService interface {
	Listen() error
	RegisterCommands() error
	Shutdown() error
}

type DiscordServiceImpl struct {
	session  *discordgo.Session
	commands []*models.DiscordCommand
}

func New(token string, shardID, shardCount int, commands []commands.Command) (*DiscordServiceImpl, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error().Err(err).Msgf("Connecting to Discord gateway failed")
		return nil, err
	}

	service := DiscordServiceImpl{
		session:  dg,
		commands: make([]*models.DiscordCommand, 0),
	}

	for _, command := range commands {
		service.commands = append(service.commands, command.GetDiscordCommand())
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

	guildId := ""
	if !viper.GetBool(models.Production) {
		log.Info().Msgf("Development mode enabled, registering commands in dedicated development guild")
		guildId = models.DevelopmentGuildId
	}

	identities := make([]*discordgo.ApplicationCommand, 0)
	for _, command := range service.commands {
		identities = append(identities, &command.Identity)
	}

	_, err := service.session.ApplicationCommandBulkOverwrite(service.session.State.User.ID, guildId, identities)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create commands, registration stopped")
		return err
	}
	log.Info().Msgf("Commands successfully registered!")

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

	if event == nil {
		return
	}

	for _, command := range service.commands {
		// TODO not always ApplicationCommandData
		if event.ApplicationCommandData().Name == command.Identity.Name {
			handler, found := command.Handlers[event.Type]
			if found {
				handler(session, event)
			} else {
				log.Error().
					Str(models.LogCommand, command.Identity.Name).
					Uint32(models.LogInteractionType, uint32(event.Type)).
					Msgf("Interaction not handled, ignoring it")
			}
		}
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
