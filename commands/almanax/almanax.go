package almanax

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(emojiService emojis.Service, requestManager requests.RequestManager) *Command {
	cmd := Command{
		emojiService:   emojiService,
		requestManager: requestManager,
	}

	subCommandHandlers := cmd.HandleSubCommand(commands.SubCommandHandlers{
		contract.AlmanaxDaySubCommandName: middlewares.
			Use(cmd.checkDate, cmd.almanaxRequest),
		contract.AlmanaxResourcesSubCommandName: middlewares.
			Use(cmd.checkDuration, cmd.resourceRequest),
		contract.AlmanaxEffectsSubCommandName: middlewares.
			Use(cmd.checkQuery, cmd.effectRequest),
	})

	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand:             subCommandHandlers,
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
	}

	return &cmd
}

func (command *Command) GetName() string {
	return contract.AlmanaxCommandName
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	command.CallHandler(s, i, lg, command.handlers)
}

func getDateOption(ctx context.Context) (*time.Time, error) {
	date, ok := ctx.Value(constants.ContextKeyDate).(*time.Time)
	if !ok {
		return nil, fmt.Errorf("cannot cast %v as *time.Time", ctx.Value(constants.ContextKeyDate))
	}

	return date, nil
}

func getDurationOption(ctx context.Context) (int32, error) {
	duration, ok := ctx.Value(constants.ContextKeyDuration).(int32)
	if !ok {
		return -1, fmt.Errorf("cannot cast %v as int32", ctx.Value(constants.ContextKeyDuration))
	}

	return duration, nil
}

func getQueryOption(ctx context.Context) (string, error) {
	query, ok := ctx.Value(constants.ContextKeyQuery).(string)
	if !ok {
		return "", fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyQuery))
	}

	return query, nil
}
