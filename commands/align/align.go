package align

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
	var minLevel float64 = constants.AlignmentMinLevel
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     slashCommandName,
			Description:              i18n.Get(constants.DefaultLocale, "align.description"),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: constants.GetDefaultPermission(),
			DMPermission:             constants.GetDMPermission(),
			DescriptionLocalizations: i18n.GetLocalizations("align.description"),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:                     getSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "align.get.description"),
					NameLocalizations:        *i18n.GetLocalizations("align.get.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("align.get.description"),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     cityOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "align.get.city.description"),
							NameLocalizations:        *i18n.GetLocalizations("align.get.city.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.get.city.description"),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 false,
							Autocomplete:             true,
						},
						{
							Name:                     orderOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "align.get.order.description"),
							NameLocalizations:        *i18n.GetLocalizations("align.get.order.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.get.order.description"),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 false,
							Autocomplete:             true,
						},
						{
							Name:                     serverOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "align.get.server.description", i18n.Vars{"game": constants.GetGame()}),
							NameLocalizations:        *i18n.GetLocalizations("align.get.server.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.get.server.description", i18n.Vars{"game": constants.GetGame()}),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 false,
							Autocomplete:             true,
						},
					},
				},
				{
					Name:                     setSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "align.set.description"),
					NameLocalizations:        *i18n.GetLocalizations("align.set.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("align.set.description"),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     cityOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "align.set.city.description"),
							NameLocalizations:        *i18n.GetLocalizations("align.set.city.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.set.city.description"),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 true,
							Autocomplete:             true,
						},
						{
							Name:                     orderOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "align.set.order.description"),
							NameLocalizations:        *i18n.GetLocalizations("align.set.order.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.set.order.description"),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 true,
							Autocomplete:             true,
						},
						{
							Name:                     levelOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "align.set.level.description"),
							NameLocalizations:        *i18n.GetLocalizations("align.set.level.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.set.level.description"),
							Type:                     discordgo.ApplicationCommandOptionInteger,
							Required:                 true,
							MinValue:                 &minLevel,
							MaxValue:                 constants.AlignmentMaxLevel,
						},
						{
							Name:                     serverOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "align.set.server.description", i18n.Vars{"game": constants.GetGame()}),
							NameLocalizations:        *i18n.GetLocalizations("align.set.server.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.set.server.description", i18n.Vars{"game": constants.GetGame()}),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 false,
							Autocomplete:             true,
						},
					},
				},
			},
		},
		Handlers: constants.DiscordHandlers{
			discordgo.InteractionApplicationCommand: middlewares.Use(command.checkCity, command.checkOrder,
				command.checkLevel, command.checkServer, command.slashRequest),
			discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
		},
	}
}

//nolint:nolintlint,exhaustive,lll,dupl
func (command *Command) GetUserCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     userCommandName,
			Type:                     discordgo.UserApplicationCommand,
			DefaultMemberPermissions: constants.GetDefaultPermission(),
			DMPermission:             constants.GetDMPermission(),
		},
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
