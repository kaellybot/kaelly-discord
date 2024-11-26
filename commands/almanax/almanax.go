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
	"github.com/spf13/viper"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(emojiService emojis.Service, requestManager requests.RequestManager) *Command {
	cmd := Command{
		AbstractCommand: commands.AbstractCommand{
			DiscordID: viper.GetString(constants.AlmanaxID),
		},
		emojiService:   emojiService,
		requestManager: requestManager,
	}

	subCommandHandlers := cmd.HandleSubCommands(commands.SubCommandHandlers{
		contract.AlmanaxDaySubCommandName: middlewares.
			Use(cmd.checkDate, cmd.getAlmanax),
		contract.AlmanaxResourcesSubCommandName: middlewares.
			Use(cmd.checkDuration, cmd.getResources),
		contract.AlmanaxEffectsSubCommandName: middlewares.
			Use(cmd.checkQuery, cmd.getAlmanaxesByEffect),
	})

	interactionHandlers := cmd.HandleInteractionMessages(commands.InteractionMessageHandlers{
		contract.AlmanaxDayCustomID:               cmd.updateAlmanax,
		contract.AlmanaxDayChoiceCustomID:         cmd.updateAlmanaxByDate,
		contract.AlmanaxEffectCustomID:            cmd.updateAlmanaxesByEffect,
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
			Name:        fmt.Sprintf("/%v day", contract.AlmanaxCommandName),
			CommandID:   fmt.Sprintf("</%v day:%v>", contract.AlmanaxCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.day", contract.AlmanaxCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.day", contract.AlmanaxCommandName)),
		},
		{
			Name:        fmt.Sprintf("/%v effects", contract.AlmanaxCommandName),
			CommandID:   fmt.Sprintf("</%v effects:%v>", contract.AlmanaxCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.effects", contract.AlmanaxCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.effects", contract.AlmanaxCommandName)),
		},
		{
			Name:        fmt.Sprintf("/%v resources", contract.AlmanaxCommandName),
			CommandID:   fmt.Sprintf("</%v resources:%v>", contract.AlmanaxCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.resources", contract.AlmanaxCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.resources", contract.AlmanaxCommandName)),
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
