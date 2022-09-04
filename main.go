package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kaellybot/kaelly-discord/models"
	"github.com/kaellybot/kaelly-discord/services/config"
	"github.com/kaellybot/kaelly-discord/services/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		return fmt.Sprintf("%s:%d", short, line)
	}
	log.Logger = log.With().Caller().Logger()

	for locale, file := range models.TranslationFiles {
		if err := i18n.LoadBundle(locale, file); err != nil {
			log.Warn().Err(err).
				Str(models.LogLocale, locale.String()).
				Str(models.LogFileName, file).
				Msgf("Cannot load translation file, continue...")
		}
	}
}

func main() {
	configService, err := config.New()
	if err != nil {
		log.Fatal().Msgf("Config service instanciation failed, shutting down.")
	}

	discordService, err := discord.New(
		configService.GetString(models.Token),
		configService.GetInt(models.ShardId),
		configService.GetInt(models.ShardCount))
	if err != nil {
		log.Fatal().Msgf("Discord service instanciation failed, shutting down.")
	}

	err = discordService.Listen()
	if err != nil {
		log.Fatal().Msgf("Discord service failed to listen events, shutting down.")
	}

	err = discordService.RegisterCommands()
	if err != nil {
		log.Fatal().Msgf("Discord service failed to register commands, shutting down.")
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	log.Info().Msgf("%s v%s is now running. Press CTRL-C to exit.", models.Name, models.Version)
	<-sc

	log.Info().Msgf("Gracefully shutting down %s...", models.Name)
	discordService.Shutdown()
}
