package almanax

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/rs/zerolog/log"
)

func (command *Command) autocomplete(s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {
	data := i.ApplicationCommandData()

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Focused {
				switch option.Name {
				case contract.AlmanaxEffectOptionName:
					command.requestAlmanaxEffectList(s, i, option.StringValue(), lg)
				default:
					log.Error().
						Str(constants.LogCommandOption, option.Name).
						Msgf("Option name not handled, ignoring it")
				}
				break
			}
		}
	}
}

func (command *Command) requestAlmanaxEffectList(s *discordgo.Session,
	i *discordgo.InteractionCreate, query string, lg discordgo.Locale) {
	if len(strings.TrimSpace(query)) == 0 {
		return
	}

	msg := mappers.MapAlmanaxEffectListRequest(query, lg)
	err := command.requestManager.Request(s, i, almanaxRequestRoutingKey,
		msg, command.autocompleteAlmanaxEffectList)
	if err != nil {
		log.Error().Err(err).Msg("Autocomplete request ignored")
	}
}

func (command *Command) autocompleteAlmanaxEffectList(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, _ map[string]any) {
	var choices []*discordgo.ApplicationCommandOptionChoice
	if isAlmanaxEffectListAnswerValid(message) {
		for _, effect := range message.EncyclopediaListAnswer.Items {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  effect.Name,
				Value: effect.Name,
			})
		}
	} else {
		log.Error().Err(commands.ErrInvalidAnswerMessage).
			Msgf("Cannot retrieve autocomplete, retrieving empty choices...")
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices[:25], // TODO too much choices
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Autocomplete request ignored")
	}
}

func isAlmanaxEffectListAnswerValid(message *amqp.RabbitMQMessage) bool {
	return message.Status == amqp.RabbitMQMessage_SUCCESS &&
		message.Type == amqp.RabbitMQMessage_ENCYCLOPEDIA_LIST_ANSWER &&
		message.EncyclopediaListAnswer != nil
}