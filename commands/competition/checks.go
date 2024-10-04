package competition

import (
	"context"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkOptionalMapNumber(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	var mapNumber int64 = 0
	for _, option := range data.Options {
		if option.Name == contract.MapNumberOptionName {
			desiredMapNumber := option.IntValue()

			if desiredMapNumber >= constants.MapNumberMin && desiredMapNumber <= constants.MapNumberMax {
				mapNumber = desiredMapNumber
				break
			} else {
				content := i18n.Get(i.Locale, "checks.map.constraints",
					i18n.Vars{"min": constants.MapNumberMin, "max": constants.MapNumberMax})
				_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
				if err != nil {
					log.Error().Err(err).Msg("Map number check response ignored")
				}
				return
			}
		}
	}

	next(context.WithValue(ctx, constants.ContextKeyMap, mapNumber))
}
