package pos

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *PosCommand) checkDimension(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, option := range data.Options {
		if option.Name == dimensionOptionName {
			dimensions := command.dimensionService.FindDimensions(option.StringValue(), lg)

			if len(dimensions) == 1 || len(strings.TrimSpace(option.StringValue())) == 0 {
				next()
			} else {
				data := discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}

				if len(dimensions) > 1 {
					data.Content = i18n.Get(lg, "pos.dimension.check.too_many", i18n.Vars{"value": option.StringValue(), "dimensions": dimensions})
				} else {
					data.Content = i18n.Get(lg, "pos.dimension.check.not_found", i18n.Vars{"value": option.StringValue()})
				}

				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &data,
				})
				if err != nil {
					log.Error().Err(err).Msg("Dimension check response ignored")
				}
			}

			break
		}
	}
}

func (command *PosCommand) checkServer(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	// TODO

	log.Info().Msgf("Check server")
	next()
}
