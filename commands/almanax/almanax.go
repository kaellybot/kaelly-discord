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
	i18n "github.com/kaysoro/discordgo-i18n"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(emojiService emojis.Service, requestManager requests.RequestManager) *Command {
	cmd := Command{
		emojiService:   emojiService,
		requestManager: requestManager,
	}

	subCommandHandlers := cmd.HandleSubCommands(commands.SubCommandHandlers{
		contract.AlmanaxDaySubCommandName: middlewares.
			Use(cmd.checkDate, cmd.getAlmanax),
		contract.AlmanaxResourcesSubCommandName: middlewares.
			Use(cmd.checkDuration, cmd.getResources),
		contract.AlmanaxEffectsSubCommandName: middlewares.
			Use(cmd.checkQuery, cmd.getAlmanaxWithEffect),
	})

	interactionHandlers := cmd.HandleInteractionMessages(commands.InteractionMessageHandlers{
		contract.AlmanaxDayCustomID:               cmd.updateAlmanax,
		contract.AlmanaxResourceCharacterCustomID: cmd.updateResourceCharacter,
		contract.AlmanaxResourceDurationCustomID:  cmd.updateResourceDuration,
	})

	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand:             subCommandHandlers,
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
		discordgo.InteractionMessageComponent:               interactionHandlers,
	}

	return &cmd
}

func (command *Command) GetName() string {
	return contract.AlmanaxCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        "/almanax day",
			CommandID:   "</almanax day:1177674483876761610>",
			Description: i18n.Get(lg, "almanax.help.detailed.day"),
			TutorialURL: i18n.Get(lg, "almanax.help.tutorial.day"),
		},
		{
			Name:        "/almanax effects",
			CommandID:   "</almanax effects:1177674483876761610>",
			Description: i18n.Get(lg, "almanax.help.detailed.effects"),
			TutorialURL: i18n.Get(lg, "almanax.help.tutorial.effects"),
		},
		{
			Name:        "/almanax resources",
			CommandID:   "</almanax resources:1177674483876761610>",
			Description: i18n.Get(lg, "almanax.help.detailed.resources"),
			TutorialURL: i18n.Get(lg, "almanax.help.tutorial.resources"),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return command.matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command.CallHandler(s, i, command.handlers)
}

func getDateOption(ctx context.Context) (*time.Time, error) {
	date, ok := ctx.Value(constants.ContextKeyDate).(*time.Time)
	if !ok {
		return nil, fmt.Errorf("cannot cast %v as *time.Time", ctx.Value(constants.ContextKeyDate))
	}

	return date, nil
}

func getDurationOption(ctx context.Context) (int64, error) {
	duration, ok := ctx.Value(constants.ContextKeyDuration).(int64)
	if !ok {
		return -1, fmt.Errorf("cannot cast %v as int64", ctx.Value(constants.ContextKeyDuration))
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

func (command *Command) matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToAlmanax(i.MessageComponentData().CustomID)
}
