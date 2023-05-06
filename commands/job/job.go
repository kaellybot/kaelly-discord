package job

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
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
	return &Command{
		bookService:    bookService,
		guildService:   guildService,
		serverService:  serverService,
		requestManager: requestManager,
	}
}

//nolint:nolintlint,exhaustive,lll,dupl
func (command *Command) GetSlashCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Handlers: constants.DiscordHandlers{
			discordgo.InteractionApplicationCommand: middlewares.Use(command.checkJob, command.checkLevel,
				command.checkServer, command.slashRequest),
			discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
		},
	}
}

//nolint:nolintlint,exhaustive,lll,dupl
func (command *Command) GetUserCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Handlers: constants.DiscordHandlers{
			discordgo.InteractionApplicationCommand: middlewares.Use(command.checkServer, command.userRequest),
		},
	}
}

func (command *Command) slashRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, _ middlewares.NextFunc) {
	for _, subCommand := range i.ApplicationCommandData().Options {
		switch subCommand.Name {
		case getSubCommandName:
			command.getRequest(ctx, s, i, lg)
		case setSubCommandName:
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
	job, ok := ctx.Value(jobOptionName).(entities.Job)
	if !ok {
		return entities.Job{}, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Job", ctx.Value(jobOptionName))
	}

	server, ok := ctx.Value(serverOptionName).(entities.Server)
	if !ok {
		return entities.Job{}, entities.Server{},
			fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(serverOptionName))
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
