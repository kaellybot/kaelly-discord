package competition

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//nolint:exhaustive // only the useful interaction types are handled.
func New(emojiService emojis.Service, requestManager requests.RequestManager,
) *Command {
	cmd := Command{
		DiscordID:      viper.GetString(constants.MapID),
		requestManager: requestManager,
		emojiService:   emojiService,
	}

	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.
			Use(cmd.checkOptionalMapNumber, cmd.getMap),
		discordgo.InteractionMessageComponent: cmd.updateMap,
	}

	return &cmd
}

func (command *Command) GetName() string {
	return contract.MapCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        fmt.Sprintf("/%v", contract.MapCommandName),
			CommandID:   fmt.Sprintf("</%v:%v>", contract.MapCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed", contract.MapCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial", contract.MapCommandName)),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return command.matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return command.CallHandler(s, i, command.handlers)
}

func (command *Command) getMap(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) error {
	mapNumber, err := getOptions(ctx)
	if err != nil {
		return err
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapCompetitionMapRequest(mapNumber, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.CompetitionRequestRoutingKey,
		msg, command.getMapReply)
}

func (command *Command) updateMap(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	customID := i.MessageComponentData().CustomID
	properties := make(map[string]any)
	var mapNumber int64
	var ok bool
	if mapNumber, ok = contract.ExtractMapNormalCustomID(customID); ok {
		properties[mapTypeProperty] = constants.MapTypeNormal
	} else if mapNumber, ok = contract.ExtractMapTacticalCustomID(customID); ok {
		properties[mapTypeProperty] = constants.MapTypeTactical
	} else {
		return fmt.Errorf("%w: custom ID %q", commands.ErrInvalidInteraction, customID)
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapCompetitionMapRequest(mapNumber, authorID, i.Locale)
	return command.requestManager.Request(s, i, constants.CompetitionRequestRoutingKey,
		msg, command.updateMapReply, properties)
}

func (command *Command) getMapReply(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) error {
	return command.updateMapReply(ctx, s, i, message,
		map[string]any{mapTypeProperty: constants.MapTypeNormal})
}

func (command *Command) updateMapReply(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) error {
	if !isAnswerValid(message) {
		return fmt.Errorf("%w: %v", commands.ErrInvalidAnswerMessage, message.GetStatus())
	}

	mapType, ok := properties[mapTypeProperty].(constants.MapType)
	if !ok {
		return fmt.Errorf("%w: %v", commands.ErrRequestPropertyNotFound, mapTypeProperty)
	}

	reply := mappers.MapCompetitionMapToWebhookEdit(message.GetCompetitionMapAnswer(),
		mapType, command.emojiService, message.Language)
	_, err := s.InteractionResponseEdit(i.Interaction, reply)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Cannot respond to interaction after receiving internal answer, ignoring request")
	}

	return nil
}

func getOptions(ctx context.Context) (int64, error) {
	mapNumber, ok := ctx.Value(constants.ContextKeyMap).(int64)
	if !ok {
		return -1,
			fmt.Errorf("cannot cast %v as int64", ctx.Value(constants.ContextKeyMap))
	}

	return mapNumber, nil
}

func isAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_COMPETITION_MAP_ANSWER &&
		message.CompetitionMapAnswer != nil
}

func (command *Command) matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToMap(i.MessageComponentData().CustomID)
}
