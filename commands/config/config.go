package config

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/kaellybot/kaelly-discord/utils/translators"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func New(guildService guilds.Service, feedService feeds.Service,
	serverService servers.Service, requestManager requests.RequestManager) *Command {
	return &Command{
		guildService:   guildService,
		feedService:    feedService,
		serverService:  serverService,
		requestManager: requestManager,
	}
}

//nolint:nolintlint,exhaustive,lll,dupl
func (command *Command) GetSlashCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     commandName,
			Description:              i18n.Get(constants.DefaultLocale, "config.description", i18n.Vars{"game": constants.GetGame()}),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: constants.GetManageServerPermission(),
			DMPermission:             constants.GetDMPermission(),
			DescriptionLocalizations: i18n.GetLocalizations("config.description", i18n.Vars{"game": constants.GetGame()}),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:                     getSubCommandName,
					Description:              i18n.Get(constants.DefaultLocale, "config.get.description"),
					NameLocalizations:        *i18n.GetLocalizations("config.get.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("config.get.description"),
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
							Name:                     languageOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.almanax.language.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.almanax.language.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.almanax.language.description"),
							Type:                     discordgo.ApplicationCommandOptionInteger,
							Required:                 false,
							Choices:                  translators.GetLocalChoices(),
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
							Name:                     feedTypeOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.rss.feedtype.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.rss.feedtype.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.rss.feedtype.description"),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 true,
							Autocomplete:             true,
						},
						{
							Name:                     languageOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.rss.language.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.rss.language.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.rss.language.description"),
							Type:                     discordgo.ApplicationCommandOptionInteger,
							Required:                 false,
							Choices:                  translators.GetLocalChoices(),
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
					Description:              i18n.Get(constants.DefaultLocale, "config.server.description", i18n.Vars{"game": constants.GetGame()}),
					NameLocalizations:        *i18n.GetLocalizations("config.server.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("config.server.description", i18n.Vars{"game": constants.GetGame()}),
					Type:                     discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:                     serverOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.server.server.description", i18n.Vars{"game": constants.GetGame()}),
							NameLocalizations:        *i18n.GetLocalizations("config.server.server.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.server.server.description", i18n.Vars{"game": constants.GetGame()}),
							Type:                     discordgo.ApplicationCommandOptionString,
							Required:                 true,
							Autocomplete:             true,
						},
						{
							Name:                     channelOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.server.channel.description", i18n.Vars{"game": constants.GetGame()}),
							NameLocalizations:        *i18n.GetLocalizations("config.server.channel.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.server.channel.description", i18n.Vars{"game": constants.GetGame()}),
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
							Name:                     languageOptionName,
							Description:              i18n.Get(constants.DefaultLocale, "config.twitter.language.description"),
							NameLocalizations:        *i18n.GetLocalizations("config.twitter.language.name"),
							DescriptionLocalizations: *i18n.GetLocalizations("config.twitter.language.description"),
							Type:                     discordgo.ApplicationCommandOptionInteger,
							Required:                 false,
							Choices:                  translators.GetLocalChoices(),
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
			discordgo.InteractionApplicationCommand: middlewares.Use(command.checkServer, command.checkEnabled,
				command.checkFeedType, command.checkLanguage, command.checkChannelID, command.request),
			discordgo.InteractionApplicationCommandAutocomplete: command.autocomplete,
		},
	}
}

func (command *Command) request(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale, _ middlewares.NextFunc) {
	for _, subCommand := range i.ApplicationCommandData().Options {
		switch subCommand.Name {
		case getSubCommandName:
			command.getRequest(ctx, s, i, lg)
		case almanaxSubCommandName:
			command.almanaxRequest(ctx, s, i, lg)
		case rssSubCommandName:
			command.rssRequest(ctx, s, i, lg)
		case twitterSubCommandName:
			command.twitterRequest(ctx, s, i, lg)
		case serverSubCommandName:
			command.serverRequest(ctx, s, i, lg)
		default:
			panic(fmt.Errorf("cannot handle subCommand %v, request ignored", subCommand.Name))
		}
	}
}

func (command *Command) createWebhook(s *discordgo.Session, channelID string) (*discordgo.Webhook, error) {
	return s.WebhookCreate(channelID, constants.Name, constants.AvatarIcon)
}

func (command *Command) getServerOptions(ctx context.Context) (entities.Server, string, error) {
	server, ok := ctx.Value(constants.ContextKeyServer).(entities.Server)
	if !ok {
		return entities.Server{}, "",
			fmt.Errorf("cannot cast %v as entities.Server", ctx.Value(constants.ContextKeyServer))
	}

	channelID := ""
	if ctx.Value(constants.ContextKeyChannel) != nil {
		channelID, ok = ctx.Value(constants.ContextKeyChannel).(string)
		if !ok {
			return entities.Server{}, "",
				fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
		}
	}

	return server, channelID, nil
}

func (command *Command) getWebhookAlmanaxOptions(ctx context.Context) (string, bool, amqp.Language, error) {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		return "", false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
	}

	enabled, ok := ctx.Value(constants.ContextKeyEnabled).(bool)
	if !ok {
		return "", false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyEnabled))
	}

	locale, ok := ctx.Value(constants.ContextKeyLanguage).(amqp.Language)
	if !ok {
		return "", false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyLanguage))
	}

	return channelID, enabled, locale, nil
}

func (command *Command) getWebhookTwitterOptions(ctx context.Context) (string, bool, amqp.Language, error) {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		return "", false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
	}

	enabled, ok := ctx.Value(constants.ContextKeyEnabled).(bool)
	if !ok {
		return "", false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyEnabled))
	}

	locale, ok := ctx.Value(constants.ContextKeyLanguage).(amqp.Language)
	if !ok {
		return "", false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyLanguage))
	}

	return channelID, enabled, locale, nil
}

func (command *Command) getWebhookRssOptions(ctx context.Context) (
	string, entities.FeedType, bool, amqp.Language, error) {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		return "", entities.FeedType{}, false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
	}

	feed, ok := ctx.Value(constants.ContextKeyFeed).(entities.FeedType)
	if !ok {
		return "", entities.FeedType{}, false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as entities.FeedType", ctx.Value(constants.ContextKeyFeed))
	}

	enabled, ok := ctx.Value(constants.ContextKeyEnabled).(bool)
	if !ok {
		return "", entities.FeedType{}, false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyEnabled))
	}

	locale, ok := ctx.Value(constants.ContextKeyLanguage).(amqp.Language)
	if !ok {
		return "", entities.FeedType{}, false, amqp.Language_ANY,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyLanguage))
	}

	return channelID, feed, enabled, locale, nil
}
