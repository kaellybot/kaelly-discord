package config

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/commands"
)

func (command *ConfigCommand) rssRequest(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, lg discordgo.Locale) {

	err := commands.DeferInteraction(s, i)
	if err != nil {
		panic(err)
	}

	/** TODO
	msg := mappers.MapConfigurationRssRequest(i.Interaction.GuildID, lg)
	err = command.requestManager.Request(s, i, configurationRequestRoutingKey, msg, command.rssRespond)
	if err != nil {
		panic(err)
	}
	**/
	content := "rss config"
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		panic(err)
	}
}
