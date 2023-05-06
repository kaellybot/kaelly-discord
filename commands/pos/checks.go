package pos

import (
	"context"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkDimension(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] dimension
	for _, option := range data.Options {
		if option.Name == contract.PosDimensionOptionName {
			dimensions := command.portalService.FindDimensions(option.StringValue(), lg)
			response, checkSuccess := validators.ExpectOnlyOneElement("checks.dimension", option.StringValue(), dimensions, lg)
			if checkSuccess {
				next(context.WithValue(ctx, constants.ContextKeyDimension, dimensions[0]))
			} else {
				_, err := s.InteractionResponseEdit(i.Interaction, &response)
				if err != nil {
					log.Error().Err(err).Msg("Dimension check response ignored")
				}
			}

			return
		}
	}

	// Option not filled, ANY dimension is then retrieved
	next(ctx)
}
