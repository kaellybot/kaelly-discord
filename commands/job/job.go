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

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(bookService books.Service, guildService guilds.Service,
	serverService servers.Service, requestManager requests.RequestManager) *Command {
	cmd := Command{
		bookService:    bookService,
		guildService:   guildService,
		serverService:  serverService,
		requestManager: requestManager,
	}

	subCommandHandlers := cmd.HandleSubCommand(commands.SubCommandHandlers{
		contract.JobGetSubCommandName: middlewares.
			Use(cmd.checkJob, cmd.checkServer, cmd.getRequest),
		contract.JobSetSubCommandName: middlewares.
			Use(cmd.checkJob, cmd.checkLevel, cmd.checkServer, cmd.setRequest),
	})

	cmd.slashHandlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand:             subCommandHandlers,
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
	}
	cmd.userHandlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.Use(cmd.checkServer, cmd.userRequest),
	}
	return &cmd
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
