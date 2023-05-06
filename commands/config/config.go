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
