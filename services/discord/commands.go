package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	"github.com/kaellybot/kaelly-discord/services/discord/commands"
)

type DiscordCommand struct {
	Identity discordgo.ApplicationCommand
	Handler  DiscordHandler
}

type DiscordHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

var (
	dmPermission            = models.DMPermission
	defaultPermission int64 = models.DefaultPermission

	discordCommands = map[string]DiscordCommand{
		commands.CommandNameAbout: {
			Identity: discordgo.ApplicationCommand{
				Name:                     commands.CommandNameAbout,
				Description:              "Gives information about Kaelly and a way to get help",
				Type:                     discordgo.ChatApplicationCommand,
				DefaultMemberPermissions: &defaultPermission,
				DMPermission:             &dmPermission,
				DescriptionLocalizations: &map[discordgo.Locale]string{
					discordgo.French: "Donne des informations sur Kaelly et un moyen d'obtenir de l'aide",
				},
			},
			Handler: commands.About,
		},
	}
)
