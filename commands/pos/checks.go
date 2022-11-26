package pos

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/rs/zerolog/log"
)

func (command *PosCommand) checkDimension(s *discordgo.Session, i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	// TODO
	log.Info().Msgf("Check dimension")
	next()
}

func (command *PosCommand) checkServer(s *discordgo.Session, i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	// TODO
	log.Info().Msgf("Check server")
	next()
}
