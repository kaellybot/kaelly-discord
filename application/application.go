package application

import (
	"errors"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/commands/about"
	"github.com/kaellybot/kaelly-discord/commands/pos"
	"github.com/kaellybot/kaelly-discord/models"
	"github.com/kaellybot/kaelly-discord/services/dimensions"
	"github.com/kaellybot/kaelly-discord/services/discord"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	ErrCannotInstantiateApp = errors.New("Cannot instantiate application")
)

type ApplicationInterface interface {
	Run() error
	Shutdown()
}

type Application struct {
	broker           amqp.MessageBrokerInterface
	guildService     guilds.GuildService
	dimensionService dimensions.DimensionService
	serverService    servers.ServerService
	discordService   discord.DiscordService
	commands         []commands.Command
	requestManager   requests.RequestManager
}

func New() (*Application, error) {
	broker, err := amqp.New(models.GetRabbitMQClientId(), viper.GetString(models.RabbitMqAddress), getBindings())
	if err != nil {
		log.Fatal().Err(err).Msgf("Broker instantiation failed, shutting down.")
	}

	requestsManager := requests.New(broker)

	guildService := guilds.New()

	dimensionService, err := dimensions.New()
	if err != nil {
		return nil, err
	}

	serverService, err := servers.New()
	if err != nil {
		return nil, err
	}

	commands := []commands.Command{
		about.New(),
		pos.New(guildService, dimensionService, serverService, requestsManager),
	}

	discordService, err := discord.New(
		viper.GetString(models.Token),
		viper.GetInt(models.ShardId),
		viper.GetInt(models.ShardCount),
		commands)
	if err != nil {
		return nil, err
	}

	return &Application{
		broker:           broker,
		requestManager:   requestsManager,
		guildService:     guildService,
		dimensionService: dimensionService,
		serverService:    serverService,
		discordService:   discordService,
		commands:         commands,
	}, nil
}

func (app *Application) Run() error {

	err := app.requestManager.Listen()
	if err != nil {
		return err
	}

	err = app.discordService.Listen()
	if err != nil {
		return err
	}

	err = app.discordService.RegisterCommands()
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) Shutdown() {
	app.broker.Shutdown()
	app.discordService.Shutdown()
	log.Info().Msgf("Application is no longer running")
}

func getBindings() []amqp.Binding {
	return []amqp.Binding{
		requests.GetBinding(),
	}
}
