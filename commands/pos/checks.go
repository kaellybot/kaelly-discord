package pos

import (
	"context"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/validators"
	i18n "github.com/kaysoro/discordgo-i18n"
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

func (command *Command) checkServer(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()

	// Filled case, expecting [1, 1] server
	for _, option := range data.Options {
		if option.Name == contract.PosServerOptionName {
			servers := command.serverService.FindServers(option.StringValue(), lg)
			response, checkSuccess := validators.ExpectOnlyOneElement("checks.server", option.StringValue(), servers, lg)
			if checkSuccess {
				next(context.WithValue(ctx, constants.ContextKeyServer, servers[0]))
			} else {
				_, err := s.InteractionResponseEdit(i.Interaction, &response)
				if err != nil {
					log.Error().Err(err).Msg("Server check response ignored")
				}
			}

			return
		}
	}

	// Option not filled (refers to guild and/or channel)
	server, found, err := command.guildService.GetServer(i.GuildID, i.ChannelID)
	if err != nil {
		panic(err)
	}

	if !found {
		content := i18n.Get(lg, "checks.server.required", i18n.Vars{"game": constants.GetGame()})
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error().Err(err).Msg("Server check response ignored")
		}
	} else {
		next(context.WithValue(ctx, constants.ContextKeyServer, server))
	}
}
