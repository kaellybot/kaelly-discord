package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/panics"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New(token string, shardID, shardCount int, slashCommands []commands.SlashCommand,
	userCommands []commands.UserCommand) (*Impl, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error().Err(err).Msgf("Connecting to Discord gateway failed")
		return nil, err
	}

	service := Impl{
		session:  dg,
		commands: make([]*constants.DiscordCommand, 0),
	}

	for _, command := range slashCommands {
		service.commands = append(service.commands, command.GetSlashCommand())
	}

	for _, command := range userCommands {
		service.commands = append(service.commands, command.GetUserCommand())
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

func (service *Impl) RegisterCommands() error {
	guildID := ""
	if !viper.GetBool(constants.Production) {
		log.Info().Msgf("Development mode enabled, registering commands in dedicated development guild")
		guildID = constants.DevelopmentGuildID
	}

	identities := make([]*discordgo.ApplicationCommand, 0)
	for _, command := range service.commands {
		identities = append(identities, &command.Identity)
	}

	_, err := service.session.ApplicationCommandBulkOverwrite(service.session.State.User.ID, guildID, identities)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create commands, registration stopped")
		return err
	}
	log.Info().Msgf("Commands successfully registered!")

	return nil
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
			Msgf("Cannot update the game status, still continuing...")
	}
}

func (service *Impl) interactionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	defer panics.HandlePanic(session, event)

	err := service.deferInteraction(session, event)
	if err != nil {
		panic(err)
	}

	locale := event.Locale
	if event.GuildLocale != nil {
		locale = *event.GuildLocale
	}

	for _, command := range service.commands {
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
			return
		}
	}
}

func (service *Impl) deferInteraction(session *discordgo.Session, event *discordgo.InteractionCreate) error {
	if event.Interaction.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
	}

	return nil
}
