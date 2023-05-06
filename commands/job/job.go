package job

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
		discordgo.InteractionApplicationCommand: middlewares.Use(command.checkJob, command.checkLevel,
			command.checkServer, command.slashRequest),
		discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
	}
	command.userHandlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.Use(command.checkServer, command.userRequest),
	}
	return &command
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return len(i.ApplicationCommandData().TargetID) == 0 &&
		contract.JobSlashCommandName == i.ApplicationCommandData().Name ||
		contract.JobUserCommandName == i.ApplicationCommandData().Name
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
		case contract.JobGetSubCommandName:
			command.getRequest(ctx, s, i, lg)
		case contract.JobSetSubCommandName:
			command.setRequest(ctx, s, i, lg)
		default:
			panic(fmt.Errorf("cannot handle subCommand %v, request ignored", subCommand.Name))
		}
	}
}

func (command *Command) userRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, _ middlewares.NextFunc) {
	command.userJobRequest(ctx, s, i, lg)
}

func (command *Command) getGetOptions(ctx context.Context) (entities.Job, entities.Server, error) {
	job, ok := ctx.Value(constants.ContextKeyJob).(entities.Job)
	if !ok {
		return entities.Job{}, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Job", ctx.Value(constants.ContextKeyJob))
	}

	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.Job{}, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	return job, server, nil
}

func (command *Command) getSetOptions(ctx context.Context) (entities.Job, int64, entities.Server, error) {
	job, ok := ctx.Value(constants.ContextKeyJob).(entities.Job)
	if !ok {
		return entities.Job{}, 0, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Job", ctx.Value(constants.ContextKeyJob))
	}

	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.Job{}, 0, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	level, ok := ctx.Value(constants.ContextKeyLevel).(int64)
	if !ok {
		return entities.Job{}, 0, entities.Server{},
			fmt.Errorf("cannot cast %v as uint", ctx.Value(constants.ContextKeyLevel))
	}

	return job, level, server, nil
}

func (command *Command) getUserOptions(ctx context.Context) (entities.Server, error) {
	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	return server, nil
}
