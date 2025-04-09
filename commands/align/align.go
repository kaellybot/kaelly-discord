package align

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/books"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/checks"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/spf13/viper"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(bookService books.Service, guildService guilds.Service,
	serverService servers.Service, emojiService emojis.Service,
	requestManager requests.RequestManager) *Command {
	cmd := Command{
		AbstractCommand: commands.AbstractCommand{
			DiscordID: viper.GetString(constants.AlignID),
		},
		bookService:    bookService,
		emojiService:   emojiService,
		guildService:   guildService,
		serverService:  serverService,
		requestManager: requestManager,
	}

	checkServer := checks.CheckServerWithFallback(contract.AlignServerOptionName,
		cmd.serverService, cmd.guildService)

	subCommandHandlers := cmd.HandleSubCommands(commands.SubCommandHandlers{
		contract.AlignGetSubCommandName: middlewares.
			Use(cmd.checkOptionalCity, cmd.checkOptionalOrder, checkServer, cmd.getBook),
		contract.AlignSetSubCommandName: middlewares.
			Use(cmd.checkMandatoryCity, cmd.checkMandatoryOrder, cmd.checkLevel, checkServer, cmd.setBook),
	})

	interactionHandlers := cmd.HandleInteractionMessages(commands.InteractionMessageHandlers{
		contract.AlignBookPageCustomID:  cmd.updateBookPage,
		contract.AlignBookCityCustomID:  cmd.updateCityBook,
		contract.AlignBookOrderCustomID: cmd.updateOrderBook,
	})

	cmd.slashHandlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand:             subCommandHandlers,
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
		discordgo.InteractionMessageComponent:               interactionHandlers,
	}
	cmd.userHandlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.Use(checkServer, cmd.userBook),
	}
	return &cmd
}

func (command *Command) GetName() string {
	return contract.AlignSlashCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        fmt.Sprintf("/%v get", contract.AlignSlashCommandName),
			CommandID:   fmt.Sprintf("</%v get:%v>", contract.AlignSlashCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.get", contract.AlignSlashCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.get", contract.AlignSlashCommandName)),
		},
		{
			Name:        fmt.Sprintf("/%v set", contract.AlignSlashCommandName),
			CommandID:   fmt.Sprintf("</%v set:%v>", contract.AlignSlashCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.set", contract.AlignSlashCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.set", contract.AlignSlashCommandName)),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return command.matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if commands.IsApplicationCommand(i) {
		if len(i.ApplicationCommandData().TargetID) == 0 {
			command.CallHandler(s, i, command.slashHandlers)
		} else {
			command.CallHandler(s, i, command.userHandlers)
		}
	} else if commands.IsMessageCommand(i) {
		command.CallHandler(s, i, command.slashHandlers)
	}
}

func getGetOptions(ctx context.Context) (entities.City, entities.Order, entities.Server, error) {
	city, ok := ctx.Value(constants.ContextKeyCity).(entities.City)
	if !ok {
		city = entities.City{}
	}

	order, ok := ctx.Value(constants.ContextKeyOrder).(entities.Order)
	if !ok {
		order = entities.Order{}
	}

	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.City{}, entities.Order{}, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	return city, order, server, nil
}

func getSetOptions(ctx context.Context) (entities.City, entities.Order,
	int64, entities.Server, error) {
	city, ok := ctx.Value(constants.ContextKeyCity).(entities.City)
	if !ok {
		return entities.City{}, entities.Order{}, 0, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.City", ctx.Value(constants.ContextKeyCity))
	}

	order, ok := ctx.Value(constants.ContextKeyOrder).(entities.Order)
	if !ok {
		return entities.City{}, entities.Order{}, 0, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Order", ctx.Value(constants.ContextKeyOrder))
	}

	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.City{}, entities.Order{}, 0, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	level, ok := ctx.Value(constants.ContextKeyLevel).(int64)
	if !ok {
		return entities.City{}, entities.Order{}, 0, entities.Server{},
			fmt.Errorf("cannot cast %v as uint", ctx.Value(constants.ContextKeyLevel))
	}

	return city, order, level, server, nil
}

func getUserOptions(ctx context.Context) (entities.Server, error) {
	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.Server{}, fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	return server, nil
}

func (command *Command) matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	if commands.IsApplicationCommand(i) {
		return len(i.ApplicationCommandData().TargetID) == 0 &&
			command.GetName() == i.ApplicationCommandData().Name ||
			contract.AlignUserCommandName == i.ApplicationCommandData().Name
	}
	return false
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToAlign(i.MessageComponentData().CustomID)
}
