package about

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New(broker amqp.MessageBroker, emojiService emojis.Service) *Command {
	return &Command{
		AbstractCommand: commands.AbstractCommand{
			DiscordID: viper.GetString(constants.AboutID),
		},
		broker:       broker,
		emojiService: emojiService,
	}
}

func (command *Command) GetName() string {
	return contract.AboutCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        fmt.Sprintf("/%v", contract.AboutCommandName),
			CommandID:   fmt.Sprintf("</%v:%v>", contract.AboutCommandName, command.DiscordID),
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.detailed", contract.AboutCommandName)),
			TutorialURL: i18n.Get(lg, fmt.Sprintf("%v.help.tutorial", contract.AboutCommandName)),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := discord.GetUserID(i.Interaction)
	message := mappers.MapAboutRequest(userID, i.Locale)
	errBroker := command.broker.Emit(message, amqp.ExchangeRequest, constants.AboutRoutingKey, i.ID)
	if errBroker != nil {
		log.Error().Err(errBroker).Msgf("Cannot trace about interaction through AMQP")
	}

	answer := mappers.MapAboutToWebhook(i.Locale, command.emojiService)
	_, err := s.InteractionResponseEdit(i.Interaction, answer)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle about reponse")
	}
}
