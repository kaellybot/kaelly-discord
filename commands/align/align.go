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
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

func New(bookService books.Service, guildService guilds.Service,
	serverService servers.Service, requestManager requests.RequestManager) *Command {
	command := Command{
		bookService:    bookService,
		guildService:   guildService,
		serverService:  serverService,
		requestManager: requestManager,
	}
	command.slashHandlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.Use(command.checkCity, command.checkOrder,
			command.checkLevel, command.checkServer, command.slashRequest),
		discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
	}
	command.userHandlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
	}
	return &command
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return len(i.ApplicationCommandData().TargetID) == 0 &&
		contract.AlignSlashCommandName == i.ApplicationCommandData().Name ||
		contract.AlignUserCommandName == i.ApplicationCommandData().Name
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	if len(i.ApplicationCommandData().TargetID) == 0 {
		command.CallHandler(s, i, lg, command.slashHandlers)
	} else {
		command.CallHandler(s, i, lg, command.userHandlers)
	}
}

func (command *Command) slashRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, _ middlewares.NextFunc) {
	for _, subCommand := range i.ApplicationCommandData().Options {
		switch subCommand.Name {
		case contract.AlignGetSubCommandName:
			command.getRequest(ctx, s, i, lg)
		case contract.AlignSetSubCommandName:
			command.setRequest(ctx, s, i, lg)
		default:
			panic(fmt.Errorf("cannot handle subCommand %v, request ignored", subCommand.Name))
		}
	}
}

func (command *Command) userRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, _ middlewares.NextFunc) {
	command.userAlignRequest(ctx, s, i, lg)
}

func (command *Command) getGetOptions(ctx context.Context) (entities.City, entities.Order, entities.Server, error) {
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

func (command *Command) getSetOptions(ctx context.Context) (entities.City, entities.Order,
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

func (command *Command) getUserOptions(ctx context.Context) (entities.Server, error) {
	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.Server{}, fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	return server, nil
}
