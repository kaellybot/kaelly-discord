package config

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func New(guildService guilds.GuildService, serverService servers.ServerService,
	requestManager requests.RequestManager) *ConfigCommand {

	return &ConfigCommand{
		guildService:   guildService,
		serverService:  serverService,
		requestManager: requestManager,
	}
}

func (command *ConfigCommand) GetDiscordCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     commandName,
			Description:              i18n.Get(constants.DefaultLocale, "config.description", i18n.Vars{"game": constants.Game}),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: &constants.DefaultPermission,
			DMPermission:             &constants.DMPermission,
			NameLocalizations:        i18n.GetLocalizations("config.name"),
			DescriptionLocalizations: i18n.GetLocalizations("config.description", i18n.Vars{"game": constants.Game}),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:                     displaySubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "config.display.description"),
					NameLocalizations:        *i18n.GetLocalizations("config.display.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("config.display.description"),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:                     almanaxSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "config.almanax.description"),
					NameLocalizations:        *i18n.GetLocalizations("config.almanax.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("config.almanax.description"),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     enabledOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.almanax.enabled.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.almanax.enabled.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.almanax.enabled.description"),
							Type:                     discordgo.ApplicationCommandOptionBoolean,
							Required:                 true,
						},
						{
							Name:                     channelOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.almanax.channel.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.almanax.channel.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.almanax.channel.description"),
							Type:                     discordgo.ApplicationCommandOptionChannel,
							Required:                 false,
						},
					},
				},
				{
					Name:                     rssSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "config.rss.description"),
					NameLocalizations:        *i18n.GetLocalizations("config.rss.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("config.rss.description"),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     enabledOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.rss.enabled.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.rss.enabled.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.rss.enabled.description"),
							Type:                     discordgo.ApplicationCommandOptionBoolean,
							Required:                 true,
						},
						{
							Name:                     channelOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.rss.channel.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.rss.channel.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.rss.channel.description"),
							Type:                     discordgo.ApplicationCommandOptionChannel,
							Required:                 false,
						},
					},
				},
				{
					Name:                     serverSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "config.server.description", i18n.Vars{"game": constants.Game}),
					NameLocalizations:        *i18n.GetLocalizations("config.server.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("config.server.description", i18n.Vars{"game": constants.Game}),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     serverOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.server.server.description", i18n.Vars{"game": constants.Game}),
							NameLocalizations:        *i18n.GetLocalizations("config.server.server.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.server.server.description", i18n.Vars{"game": constants.Game}),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 true,
							Autocomplete:             true,
						},
						{
							Name:                     channelOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.server.channel.description", i18n.Vars{"game": constants.Game}),
							NameLocalizations:        *i18n.GetLocalizations("config.server.channel.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.server.channel.description", i18n.Vars{"game": constants.Game}),
							Type:                     discordgo.ApplicationCommandOptionChannel,
							Required:                 false,
						},
					},
				},
				{
					Name:                     twitterSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "config.twitter.description"),
					NameLocalizations:        *i18n.GetLocalizations("config.twitter.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("config.twitter.description"),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     enabledOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.twitter.enabled.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.twitter.enabled.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.twitter.enabled.description"),
							Type:                     discordgo.ApplicationCommandOptionBoolean,
							Required:                 true,
						},
						{
							Name:                     channelOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.twitter.channel.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.twitter.channel.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.twitter.channel.description"),
							Type:                     discordgo.ApplicationCommandOptionChannel,
							Required:                 false,
						},
					},
				},
			},
		},
		Handlers: constants.DiscordHandlers{
			discordgo.InteractionApplicationCommand:             middlewares.Use(command.checkServer, command.request),
			discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
		},
	}
}

func (command *ConfigCommand) request(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, next middlewares.NextFunc) {

	for _, subCommand := range i.ApplicationCommandData().Options {
		switch subCommand.Name {
		case displaySubCommandName:
			command.displayRequest(ctx, s, i, lg)
		case almanaxSubCommandName:
			command.almanaxRequest(ctx, s, i, lg)
		case rssSubCommandName:
			command.rssRequest(ctx, s, i, lg)
		case twitterSubCommandName:
			command.twitterRequest(ctx, s, i, lg)
		case serverSubCommandName:
			command.serverRequest(ctx, s, i, lg)
		default:
			log.Error().Msgf("")
		}
	}
}

func (command *ConfigCommand) getServerOptions(ctx context.Context) (entities.Server, string, error) {
	server, ok := ctx.Value(serverOptionName).(entities.Server)
	if !ok {
		return entities.Server{}, "", fmt.Errorf("Cannot cast %v as entities.Server", ctx.Value(serverOptionName))
	}

	channelId := ""
	if ctx.Value(channelOptionName) != nil {
		channelId, ok = ctx.Value(channelOptionName).(string)
		if !ok {
			return entities.Server{}, "", fmt.Errorf("Cannot cast %v as string", ctx.Value(channelOptionName))
		}
	}

	return server, channelId, nil
}

func (command *ConfigCommand) getWebhookOptions(ctx context.Context) (string, bool, error) {
	channelId, ok := ctx.Value(channelOptionName).(string)
	if !ok {
		return "", false, fmt.Errorf("Cannot cast %v as string", ctx.Value(channelOptionName))
	}

	enabled, ok := ctx.Value(enabledOptionName).(bool)
	if !ok {
		return "", false, fmt.Errorf("Cannot cast %v as bool", ctx.Value(enabledOptionName))
	}

	return channelId, enabled, nil
}
