package set

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
)

func (command *Command) checkQuery(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ discordgo.Locale, next middlewares.NextFunc) {
	data := i.ApplicationCommandData()
	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Name == contract.SetQueryOptionName && len(strings.TrimSpace(option.StringValue())) == 0 {
				next(context.WithValue(ctx, constants.ContextKeyQuery, option.StringValue()))
				return
			}
		}
	}
}
