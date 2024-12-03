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
	twitterRepo "github.com/kaellybot/kaelly-discord/repositories/twitters"
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
	"github.com/kaellybot/kaelly-discord/services/twitters"
	"github.com/kaellybot/kaelly-discord/services/videasts"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/kaellybot/kaelly-discord/utils/insights"
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
	broker := amqp.New(
		constants.GetRabbitMQClientID(),
		viper.GetString(constants.RabbitMQAddress),
		amqp.WithBindings(requests.GetBinding()),
	)
	db := databases.New()
	if errDBRun := db.Run(); errDBRun != nil {
		return nil, errDBRun
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
	streamerRepo := streamerRepo.New(db)
	twitterRepo := twitterRepo.New(db)
	videastRepo := videastRepo.New(db)
	characRepo := characRepo.New(db)
	emojiRepo := emojiRepo.New(db)
	guildRepo := guildRepo.New(db)

	// Services
	portalService, errPos := portals.New(dimensionRepo, areaRepo, subAreaRepo, transportTypeRepo)
	if errPos != nil {
		return nil, errPos
	}

	serverService, errServ := servers.New(serverRepo)
	if errServ != nil {
		return nil, errServ
	}

	bookService, errBook := books.New(jobRepo, cityRepo, orderRepo)
	if errBook != nil {
		return nil, errBook
	}

	feedService, errFeed := feeds.New(feedRepo)
	if errFeed != nil {
		return nil, errFeed
	}

	streamerService, errStreamer := streamers.New(streamerRepo)
	if errStreamer != nil {
		return nil, errStreamer
	}

	twitterService, errTwitter := twitters.New(twitterRepo)
	if errTwitter != nil {
		return nil, errTwitter
	}

	videastService, errVideast := videasts.New(videastRepo)
	if errVideast != nil {
		return nil, errVideast
	}

	characService, errCharac := characteristics.New(characRepo)
	if errCharac != nil {
		return nil, errCharac
	}

	emojiService, errEmoji := emojis.New(emojiRepo)
	if errEmoji != nil {
		return nil, errEmoji
	}

	guildService := guilds.New(guildRepo)
	requestsManager := requests.New(broker)

	commands := make([]commands.DiscordCommand, 0)
	commands = append(commands,
		about.New(broker, emojiService),
		align.New(bookService, guildService, serverService, emojiService, requestsManager),
		almanax.New(emojiService, requestsManager),
		config.New(emojiService, feedService, guildService, serverService,
			streamerService, twitterService, videastService, requestsManager),
		help.New(broker, &commands),
		item.New(characService, emojiService, requestsManager),
		job.New(bookService, guildService, serverService, emojiService, requestsManager),
		competition.New(emojiService, requestsManager),
		pos.New(guildService, portalService, serverService, requestsManager),
		set.New(characService, emojiService, requestsManager),
	)

	discordService, errDiscord := discord.New(
		viper.GetString(constants.DiscordToken),
		viper.GetInt(constants.DiscordShardID),
		viper.GetInt(constants.DiscordShardCount),
		commands, guildService, emojiService, broker)
	if errDiscord != nil {
		return nil, errDiscord
	}

	// Insights
	probes := insights.NewProbes(broker.IsConnected, db.IsConnected, discordService.IsConnected)
	prom := insights.NewPrometheusMetrics()

	return &Impl{
		broker:         broker,
		db:             db,
		probes:         probes,
		prom:           prom,
		requestManager: requestsManager,
		discordService: discordService,
	}, nil
}

func (app *Impl) Run() error {
	app.probes.ListenAndServe()
	app.prom.ListenAndServe()

	if err := app.broker.Run(); err != nil {
		return err
	}

	app.requestManager.Listen()

	if err := app.discordService.Listen(); err != nil {
		return err
	}

	return nil
}

func (app *Impl) Shutdown() {
	app.discordService.Shutdown()
	app.broker.Shutdown()
	app.db.Shutdown()
	app.prom.Shutdown()
	app.probes.Shutdown()
	log.Info().Msgf("Application is no longer running")
}
