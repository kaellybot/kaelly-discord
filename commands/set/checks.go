package set

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func (command *Command) checkQuery(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, option := range data.Options {
		if option.Name == contract.SetQueryOptionName && len(strings.TrimSpace(option.StringValue())) > 0 {
			next(context.WithValue(ctx, constants.ContextKeyQuery, option.StringValue()))
			return
		}
	}

	content := i18n.Get(i.Locale, "checks.query.constraints")
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		log.Error().Err(err).Msg("Query check response ignored")
	}
}
