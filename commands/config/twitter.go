package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

func (command *ConfigCommand) twitterRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	/** TODO
	msg := mappers.MapConfigurationTwitterRequest(i.Interaction.GuildID, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.twitterRespond)
	if err != nil {
		panic(err)
	}
	**/
	content := "twitter config"
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		panic(err)
	}
}
