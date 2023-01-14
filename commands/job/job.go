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
	i18n "github.com/kaysoro/discordgo-i18n"
)

func New(bookService books.BookService, guildService guilds.GuildService,
	serverService servers.ServerService, requestManager requests.RequestManager) *JobCommand {

	return &JobCommand{
		bookService:    bookService,
		guildService:   guildService,
		serverService:  serverService,
		requestManager: requestManager,
	}
}

func (command *JobCommand) GetDiscordCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     commandName,
			Description:              i18n.Get(constants.DefaultLocale, "job.description"),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: &constants.DefaultPermission,
			DMPermission:             &constants.DMPermission,
			NameLocalizations:        i18n.GetLocalizations("job.name"),
			DescriptionLocalizations: i18n.GetLocalizations("job.description"),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:                     getSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "job.get.description"),
					NameLocalizations:        *i18n.GetLocalizations("job.get.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("job.get.description"),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     jobOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "job.get.job.description"),
							NameLocalizations:        *i18n.GetLocalizations("job.get.job.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("job.get.job.description"),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 true,
							Autocomplete:             true,
						},
						{
							Name:                     serverOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "job.get.server.description", i18n.Vars{"game": constants.Game}),
							NameLocalizations:        *i18n.GetLocalizations("job.get.server.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("job.get.server.description", i18n.Vars{"game": constants.Game}),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 false,
							Autocomplete:             true,
						},
					},
				},
				{
					Name:                     setSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "job.set.description"),
					NameLocalizations:        *i18n.GetLocalizations("job.set.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("job.set.description"),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     jobOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "job.set.job.description"),
							NameLocalizations:        *i18n.GetLocalizations("job.set.job.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("job.set.job.description"),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 true,
							Autocomplete:             true,
						},
						{
							Name:                     levelOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "job.set.level.description"),
							NameLocalizations:        *i18n.GetLocalizations("job.set.level.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("job.set.level.description"),
							Type:                     discordgo.ApplicationCommandOptionInteger,
							Required:                 true,
							MinValue:                 &minLevel,
							MaxValue:                 maxLevel,
						},
						{
							Name:                     serverOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "job.set.server.description", i18n.Vars{"game": constants.Game}),
							NameLocalizations:        *i18n.GetLocalizations("job.set.server.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("job.set.server.description", i18n.Vars{"game": constants.Game}),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 false,
							Autocomplete:             true,
						},
					},
				},
			},
		},
		Handlers: constants.DiscordHandlers{
			discordgo.InteractionApplicationCommand: middlewares.Use(command.checkJob, command.checkLevel,
				command.checkServer, command.request),
			discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
		},
	}
}

func (command *JobCommand) request(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	for _, subCommand := range i.ApplicationCommandData().Options {
		switch subCommand.Name {
		case getSubCommandName:
			command.getRequest(ctx, s, i, lg)
		case setSubCommandName:
			command.setRequest(ctx, s, i, lg)
		default:
			panic(fmt.Errorf("Cannot handle subCommand %v, request ignored", subCommand.Name))
		}
	}
}

func (command *JobCommand) getGetOptions(ctx context.Context) (entities.Job, entities.Server, error) {
	job, ok := ctx.Value(jobOptionName).(entities.Job)
	if !ok {
		return entities.Job{}, entities.Server{}, fmt.Errorf("Cannot cast %v as entities.Job", ctx.Value(jobOptionName))
	}

	server, ok := ctx.Value(serverOptionName).(entities.Server)
	if !ok {
		return entities.Job{}, entities.Server{}, fmt.Errorf("Cannot cast %v as entities.Server", ctx.Value(serverOptionName))
	}

	return job, server, nil
}

func (command *JobCommand) getSetOptions(ctx context.Context) (entities.Job, int64, entities.Server, error) {
	job, ok := ctx.Value(jobOptionName).(entities.Job)
	if !ok {
		return entities.Job{}, 0, entities.Server{}, fmt.Errorf("Cannot cast %v as entities.Job", ctx.Value(jobOptionName))
	}

	server, ok := ctx.Value(serverOptionName).(entities.Server)
	if !ok {
		return entities.Job{}, 0, entities.Server{}, fmt.Errorf("Cannot cast %v as entities.Server", ctx.Value(serverOptionName))
	}

	level, ok := ctx.Value(levelOptionName).(int64)
	if !ok {
		return entities.Job{}, 0, entities.Server{}, fmt.Errorf("Cannot cast %v as uint", ctx.Value(levelOptionName))
	}

	return job, level, server, nil
}
