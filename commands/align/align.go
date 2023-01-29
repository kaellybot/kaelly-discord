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

func New(bookService books.BookService, guildService guilds.GuildService,
	serverService servers.ServerService, requestManager requests.RequestManager) *AlignCommand {

	return &AlignCommand{
		bookService:    bookService,
		guildService:   guildService,
		serverService:  serverService,
		requestManager: requestManager,
	}
}

func (command *AlignCommand) GetSlashCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     slashCommandName,
			Description:              i18n.Get(constants.DefaultLocale, "align.description"),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: &constants.DefaultPermission,
			DMPermission:             &constants.DMPermission,
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
							Description:              i18n.Get(constants.DefaultLocale, "align.get.server.description", i18n.Vars{"game": constants.Game}),
							NameLocalizations:        *i18n.GetLocalizations("align.get.server.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.get.server.description", i18n.Vars{"game": constants.Game}),
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
							MaxValue:                 maxLevel,
						},
						{
							Name:                     serverOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "align.set.server.description", i18n.Vars{"game": constants.Game}),
							NameLocalizations:        *i18n.GetLocalizations("align.set.server.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("align.set.server.description", i18n.Vars{"game": constants.Game}),
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

func (command *AlignCommand) GetUserCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     userCommandName,
			Type:                     discordgo.UserApplicationCommand,
			DefaultMemberPermissions: &constants.DefaultPermission,
			DMPermission:             &constants.DMPermission,
		},
		Handlers: constants.DiscordHandlers{
			discordgo.InteractionApplicationCommand: middlewares.Use(command.checkServer, command.userRequest),
		},
	}
}

func (command *AlignCommand) slashRequest(ctx context.Context, s *discordgo.Session,
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

func (command *AlignCommand) userRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	command.userAlignRequest(ctx, s, i, lg)
}

func (command *AlignCommand) getGetOptions(ctx context.Context) (entities.City, entities.Order, entities.Server, error) {
	city, ok := ctx.Value(cityOptionName).(entities.City)
	if !ok {
		city = entities.City{}
	}

	order, ok := ctx.Value(orderOptionName).(entities.Order)
	if !ok {
		order = entities.Order{}
	}

	server, ok := ctx.Value(serverOptionName).(entities.Server)
	if !ok {
		return entities.City{}, entities.Order{}, entities.Server{},
			fmt.Errorf("Cannot cast %v as entities.Server", ctx.Value(serverOptionName))
	}

	return city, order, server, nil
}

func (command *AlignCommand) getSetOptions(ctx context.Context) (entities.City, entities.Order,
	int64, entities.Server, error) {

	city, ok := ctx.Value(cityOptionName).(entities.City)
	if !ok {
		return entities.City{}, entities.Order{}, 0, entities.Server{},
			fmt.Errorf("Cannot cast %v as entities.City", ctx.Value(cityOptionName))
	}

	order, ok := ctx.Value(orderOptionName).(entities.Order)
	if !ok {
		return entities.City{}, entities.Order{}, 0, entities.Server{},
			fmt.Errorf("Cannot cast %v as entities.Order", ctx.Value(orderOptionName))
	}

	server, ok := ctx.Value(serverOptionName).(entities.Server)
	if !ok {
		return entities.City{}, entities.Order{}, 0, entities.Server{},
			fmt.Errorf("Cannot cast %v as entities.Server", ctx.Value(serverOptionName))
	}

	level, ok := ctx.Value(levelOptionName).(int64)
	if !ok {
		return entities.City{}, entities.Order{}, 0, entities.Server{},
			fmt.Errorf("Cannot cast %v as uint", ctx.Value(levelOptionName))
	}

	return city, order, level, server, nil
}

func (command *AlignCommand) getUserOptions(ctx context.Context) (entities.Server, error) {
	server, ok := ctx.Value(serverOptionName).(entities.Server)
	if !ok {
		return entities.Server{}, fmt.Errorf("Cannot cast %v as entities.Server", ctx.Value(serverOptionName))
	}

	return server, nil
}
