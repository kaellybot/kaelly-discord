package about

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func New() *AboutCommand {
	return &AboutCommand{}
}

func (command *AboutCommand) GetSlashCommand() *constants.DiscordCommand {
	return &constants.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     commandName,
			Description:              i18n.Get(constants.DefaultLocale, "about.description"),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: &constants.DefaultPermission,
			DMPermission:             &constants.DMPermission,
			DescriptionLocalizations: i18n.GetLocalizations("about.description"),
		},
		Handlers: constants.DiscordHandlers{
			discordgo.InteractionApplicationCommand: command.about,
		},
	}
}

func (command *AboutCommand) about(s *discordgo.Session, i *discordgo.InteractionCreate, lg discordgo.Locale) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{command.getAboutEmbed(lg)},
	})
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle about reponse")
	}
}

func (command *AboutCommand) getAboutEmbed(locale discordgo.Locale) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       i18n.Get(locale, "about.title", i18n.Vars{"name": constants.Name, "version": constants.Version}),
		Description: i18n.Get(locale, "about.desc", i18n.Vars{"game": constants.Game}),
		Color:       constants.Color,
		Image:       &discordgo.MessageEmbedImage{URL: constants.AvatarImage},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: constants.Game.Icon},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    i18n.Get(locale, "about.footer"),
			IconURL: constants.AnkamaLogo,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   i18n.Get(locale, "about.support.title"),
				Value:  i18n.Get(locale, "about.support.desc", i18n.Vars{"discord": constants.Discord}),
				Inline: false,
			},
			{
				Name:   i18n.Get(locale, "about.twitter.title"),
				Value:  i18n.Get(locale, "about.twitter.desc", i18n.Vars{"twitter": constants.Twitter}),
				Inline: false,
			},
			{
				Name:   i18n.Get(locale, "about.opensource.title"),
				Value:  i18n.Get(locale, "about.opensource.desc", i18n.Vars{"github": constants.Github}),
				Inline: false,
			},
			{
				Name:   i18n.Get(locale, "about.free.title"),
				Value:  i18n.Get(locale, "about.free.desc", i18n.Vars{"paypal": constants.Paypal}),
				Inline: false,
			},
			{
				Name:   i18n.Get(locale, "about.privacy.title"),
				Value:  i18n.Get(locale, "about.privacy.desc"),
				Inline: false,
			},
			{
				Name:   i18n.Get(locale, "about.graphist.title"),
				Value:  i18n.Get(locale, "about.graphist.desc", i18n.Vars{"graphist": constants.Elycann}),
				Inline: false,
			},
		},
	}
}
