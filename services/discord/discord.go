package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/panics"
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
	commands []*constants.DiscordCommand
}

func New(token string, shardID, shardCount int, commands []commands.Command) (*DiscordServiceImpl, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error().Err(err).Msgf("Connecting to Discord gateway failed")
		return nil, err
	}

	service := DiscordServiceImpl{
		session:  dg,
		commands: make([]*constants.DiscordCommand, 0),
	}

	for _, command := range commands {
		service.commands = append(service.commands, command.GetDiscordCommand())
	}

	dg.Identify.Shard = &[2]int{shardID, shardCount}
	dg.Identify.Intents = constants.Intents
	dg.AddHandler(service.ready)
	dg.AddHandler(service.messageCreate)
	dg.AddHandler(service.interactionCreate)

	return &service, nil
}

func (service *DiscordServiceImpl) Listen() error {
	err := service.session.Open()
	if err != nil {
		log.Error().Int(constants.LogShard, service.session.ShardID).Err(err).Msgf("Impossible to listen events")
		return err
	}

	log.Info().Int(constants.LogShard, service.session.ShardID).Msgf("Discord service is listening events...")
	return nil
}

func (service *DiscordServiceImpl) RegisterCommands() error {

	guildId := ""
	if !viper.GetBool(constants.Production) {
		log.Info().Msgf("Development mode enabled, registering commands in dedicated development guild")
		guildId = constants.DevelopmentGuildId
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
	log.Info().Int(constants.LogShard, service.session.ShardID).Msgf("Closing Discord connections...")
	return service.session.Close()
}

func (service *DiscordServiceImpl) ready(session *discordgo.Session, event *discordgo.Ready) {
	log.Info().Int(constants.LogShard, session.ShardID).Int(constants.LogGuildCount, len(session.State.Guilds)).Msgf("Ready!")
	session.UpdateGameStatus(0, constants.Game)
}

func (service *DiscordServiceImpl) messageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	// TODO defer service.handlePanic()
}

func (service *DiscordServiceImpl) interactionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	defer panics.HandlePanic(session, event)

	if event == nil {
		return
	}

	locale := event.Locale
	if event.GuildLocale != nil {
		locale = *event.GuildLocale
	}

	for _, command := range service.commands {
		// TODO not always ApplicationCommandData
		if event.ApplicationCommandData().Name == command.Identity.Name {
			handler, found := command.Handlers[event.Type]
			if found {
				handler(session, event, locale)
			} else {
				log.Error().
					Str(constants.LogCommand, command.Identity.Name).
					Uint32(constants.LogInteractionType, uint32(event.Type)).
					Msgf("Interaction not handled, ignoring it")
			}
		}
	}
}
