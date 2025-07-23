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
	"github.com/kaellybot/kaelly-discord/commands/set"
	"github.com/kaellybot/kaelly-discord/models/constants"
	almanaxRepo "github.com/kaellybot/kaelly-discord/repositories/almanaxes"
	characRepo "github.com/kaellybot/kaelly-discord/repositories/characteristics"
	"github.com/kaellybot/kaelly-discord/repositories/cities"
	emojiRepo "github.com/kaellybot/kaelly-discord/repositories/emojis"
	feedRepo "github.com/kaellybot/kaelly-discord/repositories/feeds"
	guildRepo "github.com/kaellybot/kaelly-discord/repositories/guilds"
	"github.com/kaellybot/kaelly-discord/repositories/jobs"
	"github.com/kaellybot/kaelly-discord/repositories/orders"
	serverRepo "github.com/kaellybot/kaelly-discord/repositories/servers"
	twitterRepo "github.com/kaellybot/kaelly-discord/repositories/twitters"
	"github.com/kaellybot/kaelly-discord/repositories/weapons"
	"github.com/kaellybot/kaelly-discord/services/almanaxes"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
	"github.com/kaellybot/kaelly-discord/services/discord"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/equipments"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/twitters"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/kaellybot/kaelly-discord/utils/insights"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//nolint:funlen,nolintlint // Yup, but much clearer like that.
func New() (*Impl, error) {
	if !viper.GetBool(constants.Production) {
		log.Info().Msgf("Development mode enabled")
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
	almanaxRepo := almanaxRepo.New(db)
	characRepo := characRepo.New(db)
	cityRepo := cities.New(db)
	emojiRepo := emojiRepo.New(db)
	feedRepo := feedRepo.New(db)
	guildRepo := guildRepo.New(db)
	jobRepo := jobs.New(db)
	orderRepo := orders.New(db)
	serverRepo := serverRepo.New(db)
	twitterRepo := twitterRepo.New(db)
	weaponRepo := weapons.New(db)

	// Services
	almanaxService, errAlm := almanaxes.New(almanaxRepo)
	if errAlm != nil {
		return nil, errAlm
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

	twitterService, errTwitter := twitters.New(twitterRepo)
	if errTwitter != nil {
		return nil, errTwitter
	}

	characService, errCharac := characteristics.New(characRepo)
	if errCharac != nil {
		return nil, errCharac
	}

	equipmentService, errEquip := equipments.New(weaponRepo)
	if errEquip != nil {
		return nil, errEquip
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
		config.New(almanaxService, emojiService, feedService, guildService,
			serverService, twitterService, requestsManager),
		help.New(broker, &commands, emojiService),
		item.New(characService, equipmentService, emojiService, requestsManager),
		job.New(bookService, guildService, serverService, emojiService, requestsManager),
		competition.New(emojiService, requestsManager),
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
