package application

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/commands/about"
	"github.com/kaellybot/kaelly-discord/commands/align"
	"github.com/kaellybot/kaelly-discord/commands/almanax"
	"github.com/kaellybot/kaelly-discord/commands/competition"
	"github.com/kaellybot/kaelly-discord/commands/config"
	"github.com/kaellybot/kaelly-discord/commands/help"
	"github.com/kaellybot/kaelly-discord/commands/item"
	"github.com/kaellybot/kaelly-discord/commands/job"
	"github.com/kaellybot/kaelly-discord/commands/pos"
	"github.com/kaellybot/kaelly-discord/commands/set"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/repositories/areas"
	characRepo "github.com/kaellybot/kaelly-discord/repositories/characteristics"
	"github.com/kaellybot/kaelly-discord/repositories/cities"
	"github.com/kaellybot/kaelly-discord/repositories/dimensions"
	emojiRepo "github.com/kaellybot/kaelly-discord/repositories/emojis"
	feedRepo "github.com/kaellybot/kaelly-discord/repositories/feeds"
	guildRepo "github.com/kaellybot/kaelly-discord/repositories/guilds"
	"github.com/kaellybot/kaelly-discord/repositories/jobs"
	"github.com/kaellybot/kaelly-discord/repositories/orders"
	serverRepo "github.com/kaellybot/kaelly-discord/repositories/servers"
	streamerRepo "github.com/kaellybot/kaelly-discord/repositories/streamers"
	"github.com/kaellybot/kaelly-discord/repositories/subareas"
	"github.com/kaellybot/kaelly-discord/repositories/transports"
	videastRepo "github.com/kaellybot/kaelly-discord/repositories/videasts"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/discord"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/portals"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/streamers"
	"github.com/kaellybot/kaelly-discord/services/videasts"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//nolint:funlen,nolintlint // Yup, but much clearer like that.
func New() (*Impl, error) {
	if !viper.GetBool(constants.Production) {
		log.Info().Msgf("Development mode enabled, retrieving specific values (eg. emoji_dev)")
	}

	// Misc
	db, err := databases.New()
	if err != nil {
		log.Fatal().Err(err).Msgf("DB instantiation failed, shutting down.")
	}

	broker, err := amqp.New(constants.GetRabbitMQClientID(), viper.GetString(constants.RabbitMQAddress), getBindings())
	if err != nil {
		log.Fatal().Err(err).Msgf("Broker instantiation failed, shutting down.")
	}

	// Repositories
	dimensionRepo := dimensions.New(db)
	areaRepo := areas.New(db)
	subAreaRepo := subareas.New(db)
	transportTypeRepo := transports.New(db)
	serverRepo := serverRepo.New(db)
	jobRepo := jobs.New(db)
	cityRepo := cities.New(db)
	orderRepo := orders.New(db)
	feedRepo := feedRepo.New(db)
	videastRepo := videastRepo.New(db)
	streamerRepo := streamerRepo.New(db)
	characRepo := characRepo.New(db)
	emojiRepo := emojiRepo.New(db)
	guildRepo := guildRepo.New(db)

	// Services
	portalService, err := portals.New(dimensionRepo, areaRepo, subAreaRepo, transportTypeRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Dimension Service instantiation failed, shutting down.")
	}

	serverService, err := servers.New(serverRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Server Service instantiation failed, shutting down.")
	}

	bookService, err := books.New(jobRepo, cityRepo, orderRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Book Service instantiation failed, shutting down.")
	}

	feedService, err := feeds.New(feedRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Feed Service instantiation failed, shutting down.")
	}

	videastService, err := videasts.New(videastRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Videast Service instantiation failed, shutting down.")
	}

	streamerService, err := streamers.New(streamerRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Streamer Service instantiation failed, shutting down.")
	}

	characService, err := characteristics.New(characRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Characteristic Service instantiation failed, shutting down.")
	}

	emojiService, err := emojis.New(emojiRepo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Emoji Service instantiation failed, shutting down.")
	}

	guildService := guilds.New(guildRepo)
	requestsManager := requests.New(broker)

	commands := make([]commands.DiscordCommand, 0)
	commands = append(commands,
		about.New(broker, emojiService),
		align.New(bookService, guildService, serverService, emojiService, requestsManager),
		almanax.New(emojiService, requestsManager),
		config.New(emojiService, feedService, guildService, serverService,
			streamerService, videastService, requestsManager),
		help.New(broker, &commands),
		item.New(characService, emojiService, requestsManager),
		job.New(bookService, guildService, serverService, emojiService, requestsManager),
		competition.New(emojiService, requestsManager),
		pos.New(guildService, portalService, serverService, requestsManager),
		set.New(characService, emojiService, requestsManager),
	)

	discordService, err := discord.New(
		viper.GetString(constants.Token),
		viper.GetInt(constants.ShardID),
		viper.GetInt(constants.ShardCount), commands)
	if err != nil {
		return nil, err
	}

	return &Impl{
		db:             db,
		broker:         broker,
		requestManager: requestsManager,
		discordService: discordService,
	}, nil
}

func (app *Impl) Run() error {
	err := app.requestManager.Listen()
	if err != nil {
		return err
	}

	err = app.discordService.Listen()
	if err != nil {
		return err
	}

	return nil
}

func (app *Impl) Shutdown() {
	app.db.Shutdown()
	app.broker.Shutdown()
	app.discordService.Shutdown()
	log.Info().Msgf("Application is no longer running")
}

func getBindings() []amqp.Binding {
	return []amqp.Binding{
		requests.GetBinding(),
	}
}
