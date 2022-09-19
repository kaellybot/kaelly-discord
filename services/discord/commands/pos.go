package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func Pos() *models.DiscordCommand {
	return &models.DiscordCommand{
		Identity: discordgo.ApplicationCommand{
			Name:                     i18n.Get(models.DefaultLocale, "pos.name"),
			Description:              i18n.Get(models.DefaultLocale, "pos.description"),
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: &models.DefaultPermission,
			DMPermission:             &models.DMPermission,
			NameLocalizations:        i18n.GetLocalizations("pos.name"),
			DescriptionLocalizations: i18n.GetLocalizations("pos.description"),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:                     i18n.Get(models.DefaultLocale, "pos.dimension.name"),
					Description:              i18n.Get(models.DefaultLocale, "pos.dimension.description"),
					NameLocalizations:        *i18n.GetLocalizations("pos.dimension.name"),
					DescriptionLocalizations: *i18n.GetLocalizations("pos.dimension.description"),
					Type:                     discordgo.ApplicationCommandOptionString,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:              i18n.Get(models.DefaultLocale, "dimensions.enutrosor"),
							NameLocalizations: *i18n.GetLocalizations("dimensions.enutrosor"),
							Value:             models.Enutrosor,
						},
						{
							Name:              i18n.Get(models.DefaultLocale, "dimensions.srambad"),
							NameLocalizations: *i18n.GetLocalizations("dimensions.srambad"),
							Value:             models.Srambad,
						},
						{
							Name:              i18n.Get(models.DefaultLocale, "dimensions.xelorium"),
							NameLocalizations: *i18n.GetLocalizations("dimensions.xelorium"),
							Value:             models.Xelorium,
						},
						{
							Name:              i18n.Get(models.DefaultLocale, "dimensions.ecaflipus"),
							NameLocalizations: *i18n.GetLocalizations("dimensions.ecaflipus"),
							Value:             models.Ecaflipus,
						},
					},
				},
			},
		},
		Handler: pos,
	}
}

func pos(s *discordgo.Session, i *discordgo.InteractionCreate) {

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle pos defer reponse")
	}

	// TODO send models to rabbitmq, retrieve response when possible

	data := make([]*discordgo.MessageEmbed, 0)
	if len(i.ApplicationCommandData().Options) > 0 {
		data = append(data, &discordgo.MessageEmbed{Title: i.ApplicationCommandData().Options[0].StringValue()})
	} else {
		data = append(data,
			&discordgo.MessageEmbed{Title: models.Enutrosor},
			&discordgo.MessageEmbed{Title: models.Srambad},
			&discordgo.MessageEmbed{Title: models.Xelorium},
			&discordgo.MessageEmbed{Title: models.Ecaflipus},
		)
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &data,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle pos reponse")
	}
}
