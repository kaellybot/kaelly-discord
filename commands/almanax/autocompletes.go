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
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/rs/zerolog/log"
)

func (command *Command) autocomplete(s *discordgo.Session,
	i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	for _, subCommand := range data.Options {
		for _, option := range subCommand.Options {
			if option.Focused {
				switch option.Name {
				case contract.AlmanaxEffectOptionName:
					command.requestAlmanaxEffectList(s, i, option.StringValue())
				default:
					log.Error().
						Str(constants.LogCommandOption, option.Name).
						Msgf("Option name not handled, answering without choices")
					respondAutocomplete(s, i, nil)
				}
				return
			}
		}
	}

	respondAutocomplete(s, i, nil)
}

func (command *Command) requestAlmanaxEffectList(s *discordgo.Session,
	i *discordgo.InteractionCreate, query string) {
	if len(strings.TrimSpace(query)) == 0 {
		respondAutocomplete(s, i, nil)
		return
	}

	authorID := discord.GetUserID(i.Interaction)
	msg := mappers.MapAlmanaxEffectListRequest(query, authorID, i.Locale)
	err := command.requestManager.Request(s, i, constants.AlmanaxRequestRoutingKey,
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

	respondAutocomplete(s, i, choices)
}

// respondAutocomplete answers the interaction; Discord expects a response for
// every keystroke, even when there is nothing to suggest.
func respondAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate,
	choices []*discordgo.ApplicationCommandOptionChoice) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
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
