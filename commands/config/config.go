package config

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/services/feeds"
	"github.com/kaellybot/kaelly-discord/services/guilds"
	"github.com/kaellybot/kaelly-discord/services/servers"
	"github.com/kaellybot/kaelly-discord/services/streamers"
	"github.com/kaellybot/kaelly-discord/services/twitters"
	"github.com/kaellybot/kaelly-discord/services/videasts"
	"github.com/kaellybot/kaelly-discord/utils/checks"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//nolint:exhaustive // only useful handlers must be implemented, it will panic also
func New(emojiService emojis.Service, feedService feeds.Service,
	guildService guilds.Service, serverService servers.Service,
	streamerService streamers.Service, twitterService twitters.Service,
	videastService videasts.Service, requestManager requests.RequestManager) *Command {
	cmd := Command{
		AbstractCommand: commands.AbstractCommand{
			DiscordID: viper.GetString(constants.ConfigID),
		},
		emojiService:    emojiService,
		feedService:     feedService,
		guildService:    guildService,
		serverService:   serverService,
		streamerService: streamerService,
		twitterService:  twitterService,
		videastService:  videastService,
		requestManager:  requestManager,
	}

	checkServer := checks.CheckServer(contract.ConfigServerOptionName, cmd.serverService)

	subCommandHandlers := cmd.HandleSubCommands(commands.SubCommandHandlers{
		contract.ConfigAlmanaxSubCommandName: middlewares.
			Use(cmd.checkEnabled, cmd.checkChannelID, cmd.almanaxRequest),
		contract.ConfigGetSubCommandName: middlewares.
			Use(cmd.getRequest),
		contract.ConfigRSSSubCommandName: middlewares.
			Use(cmd.checkEnabled, cmd.checkFeedType, cmd.checkChannelID, cmd.rssRequest),
		contract.ConfigServerSubCommandName: middlewares.
			Use(checkServer, cmd.checkChannelID, cmd.serverRequest),
		contract.ConfigTwitchSubCommandName: middlewares.
			Use(cmd.checkEnabled, cmd.checkStreamer, cmd.checkChannelID, cmd.twitchRequest),
		contract.ConfigTwitterSubCommandName: middlewares.
			Use(cmd.checkEnabled, cmd.checkTwitterAccount, cmd.checkChannelID, cmd.twitterRequest),
		contract.ConfigYoutubeSubCommandName: middlewares.
			Use(cmd.checkEnabled, cmd.checkVideast, cmd.checkChannelID, cmd.youtubeRequest),
	})

	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand:             subCommandHandlers,
		discordgo.InteractionApplicationCommandAutocomplete: cmd.autocomplete,
	}

	return &cmd
}

func (command *Command) GetName() string {
	return contract.ConfigCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        fmt.Sprintf("/%v get", contract.ConfigCommandName),
			CommandID:   fmt.Sprintf("</%v get:%v>", contract.ConfigCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.get", contract.ConfigCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.get", contract.ConfigCommandName)),
		},
		{
			Name:      fmt.Sprintf("/%v almanax", contract.ConfigCommandName),
			CommandID: fmt.Sprintf("</%v almanax:%v>", contract.ConfigCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.almanax", contract.ConfigCommandName),
				i18n.Vars{
					"defaultLocale": constants.DefaultLocale,
				}),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.almanax", contract.ConfigCommandName)),
		},
		{
			Name:        fmt.Sprintf("/%v rss", contract.ConfigCommandName),
			CommandID:   fmt.Sprintf("</%v rss:%v>", contract.ConfigCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.rss", contract.ConfigCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.rss", contract.ConfigCommandName)),
		},
		{
			Name:        fmt.Sprintf("/%v server", contract.ConfigCommandName),
			CommandID:   fmt.Sprintf("</%v server:%v>", contract.ConfigCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.server", contract.ConfigCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.server", contract.ConfigCommandName)),
		},
		{
			Name:      fmt.Sprintf("/%v twitch", contract.ConfigCommandName),
			CommandID: fmt.Sprintf("</%v twitch:%v>", contract.ConfigCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.twitch", contract.ConfigCommandName),
				i18n.Vars{
					"discord": constants.Discord,
				}),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.twitch", contract.ConfigCommandName)),
		},
		{
			Name:        fmt.Sprintf("/%v twitter", contract.ConfigCommandName),
			CommandID:   fmt.Sprintf("</%v twitter:%v>", contract.ConfigCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.twitter", contract.ConfigCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.twitter", contract.ConfigCommandName)),
		},
		{
			Name:      fmt.Sprintf("/%v youtube", contract.ConfigCommandName),
			CommandID: fmt.Sprintf("</%v youtube:%v>", contract.ConfigCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed.youtube", contract.ConfigCommandName),
				i18n.Vars{
					"discord": constants.Discord,
				}),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial.youtube", contract.ConfigCommandName)),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command.CallHandler(s, i, command.handlers)
}

func (command *Command) createWebhook(s *discordgo.Session, i *discordgo.InteractionCreate,
	channelID string) (*discordgo.Webhook, bool) {
	webhook, err := s.WebhookCreate(channelID, constants.Name, constants.AvatarWebhook)
	if err != nil {
		apiError, ok := discord.ExtractAPIError(err)
		if ok || apiError.Code == constants.DiscordCodeTooManyWebhooks {
			content := i18n.Get(i.Locale, "errors.too_many_webhooks")
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			if err != nil {
				log.Error().Err(err).Msg("Too many webhooks error response ignored")
			}
			return nil, false
		}

		panic(err)
	}

	return webhook, true
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

func getWebhookAlmanaxOptions(ctx context.Context) (string, bool, error) {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		return "", false,
			fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
	}

	enabled, ok := ctx.Value(constants.ContextKeyEnabled).(bool)
	if !ok {
		return "", false,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyEnabled))
	}

	return channelID, enabled, nil
}

func getWebhookTwitchOptions(ctx context.Context) (
	string, entities.Streamer, bool, error) {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		return "", entities.Streamer{}, false,
			fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
	}

	streamer, ok := ctx.Value(constants.ContextKeyStreamer).(entities.Streamer)
	if !ok {
		return "", entities.Streamer{}, false,
			fmt.Errorf("cannot cast %v as entities.Streamer", ctx.Value(constants.ContextKeyStreamer))
	}

	enabled, ok := ctx.Value(constants.ContextKeyEnabled).(bool)
	if !ok {
		return "", entities.Streamer{}, false,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyEnabled))
	}

	return channelID, streamer, enabled, nil
}

func getWebhookTwitterOptions(ctx context.Context) (string, entities.TwitterAccount, bool, error) {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		return "", entities.TwitterAccount{}, false,
			fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
	}

	twitterAccount, ok := ctx.Value(constants.ContextKeyTwitter).(entities.TwitterAccount)
	if !ok {
		return "", entities.TwitterAccount{}, false,
			fmt.Errorf("cannot cast %v as entities.TwitterAccount", ctx.Value(constants.ContextKeyTwitter))
	}

	enabled, ok := ctx.Value(constants.ContextKeyEnabled).(bool)
	if !ok {
		return "", entities.TwitterAccount{}, false,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyEnabled))
	}

	return channelID, twitterAccount, enabled, nil
}

func getWebhookRssOptions(ctx context.Context) (
	string, entities.FeedType, bool, error) {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		return "", entities.FeedType{}, false,
			fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
	}

	feed, ok := ctx.Value(constants.ContextKeyFeed).(entities.FeedType)
	if !ok {
		return "", entities.FeedType{}, false,
			fmt.Errorf("cannot cast %v as entities.FeedType", ctx.Value(constants.ContextKeyFeed))
	}

	enabled, ok := ctx.Value(constants.ContextKeyEnabled).(bool)
	if !ok {
		return "", entities.FeedType{}, false,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyEnabled))
	}

	return channelID, feed, enabled, nil
}

func getWebhookYoutubeOptions(ctx context.Context) (
	string, entities.Videast, bool, error) {
	channelID, ok := ctx.Value(constants.ContextKeyChannel).(string)
	if !ok {
		return "", entities.Videast{}, false,
			fmt.Errorf("cannot cast %v as string", ctx.Value(constants.ContextKeyChannel))
	}

	videast, ok := ctx.Value(constants.ContextKeyVideast).(entities.Videast)
	if !ok {
		return "", entities.Videast{}, false,
			fmt.Errorf("cannot cast %v as entities.Videast", ctx.Value(constants.ContextKeyVideast))
	}

	enabled, ok := ctx.Value(constants.ContextKeyEnabled).(bool)
	if !ok {
		return "", entities.Videast{}, false,
			fmt.Errorf("cannot cast %v as bool", ctx.Value(constants.ContextKeyEnabled))
	}

	return channelID, videast, enabled, nil
}
