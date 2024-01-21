package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/rs/zerolog/log"
)

func (command *AbstractCommand) CallHandler(s *discordgo.Session, i *discordgo.InteractionCreate,
	handlers DiscordHandlers) {
	handler, found := handlers[i.Type]
	if found {
		handler(s, i)
	} else {
		log.Error().
			Uint32(constants.LogInteractionType, uint32(i.Type)).
			Msgf("Interaction not handled, ignoring it")
	}
}

func (command *AbstractCommand) HandleSubCommand(handlers map[string]DiscordHandler) DiscordHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if IsApplicationCommand(i) {
			data := i.ApplicationCommandData()
			for _, subCommand := range data.Options {
				handler, found := handlers[subCommand.Name]
				if found {
					handler(s, i)
				} else {
					panic(ErrNoSubCommandHandler)
				}
			}
		}
	}
}

func IsApplicationCommand(i *discordgo.InteractionCreate) bool {
	return i.Type == discordgo.InteractionApplicationCommand ||
		i.Type == discordgo.InteractionApplicationCommandAutocomplete
}

func IsMessageCommand(i *discordgo.InteractionCreate) bool {
	return i.Type == discordgo.InteractionMessageComponent
}
