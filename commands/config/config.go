package config

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/utils/checks"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(guildService guilds.Service, feedService feeds.Service,
	serverService servers.Service, requestManager requests.RequestManager) *Command {
	cmd := Command{
		guildService:   guildService,
		feedService:    feedService,
		serverService:  serverService,
		requestManager: requestManager,
	}

	checkServer := checks.CheckServer(contract.ConfigServerOptionName, cmd.serverService)

	subCommandHandlers := cmd.HandleSubCommand(commands.SubCommandHandlers{
		contract.ConfigAlmanaxSubCommandName: middlewares.
			Use(cmd.checkEnabled, cmd.checkLanguage, cmd.checkChannelID, cmd.almanaxRequest),
		contract.ConfigGetSubCommandName: middlewares.
			Use(cmd.getRequest),
		contract.ConfigRSSSubCommandName: middlewares.
			Use(cmd.checkEnabled, cmd.checkFeedType, cmd.checkLanguage, cmd.checkChannelID, cmd.rssRequest),
		contract.ConfigServerSubCommandName: middlewares.
			Use(checkServer, cmd.checkChannelID, cmd.serverRequest),
		contract.ConfigTwitterSubCommandName: middlewares.
			Use(cmd.checkEnabled, cmd.checkLanguage, cmd.checkChannelID, cmd.twitterRequest),
	})

	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand:             subCommandHandlers,
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
	}

	return &cmd
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == contract.ConfigCommandName
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	command.CallHandler(s, i, lg, command.handlers)
}

func (command *Command) createWebhook(s *discordgo.Session, channelID string) (*discordgo.Webhook, error) {
	return s.WebhookCreate(channelID, constants.Name, constants.AvatarIcon)
}

func getServerOptions(ctx context.Context) (entities.Server, string, error) {
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

func getWebhookAlmanaxOptions(ctx context.Context) (string, bool, amqp.Language, error) {
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

func getWebhookTwitterOptions(ctx context.Context) (string, bool, amqp.Language, error) {
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

func getWebhookRssOptions(ctx context.Context) (
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
