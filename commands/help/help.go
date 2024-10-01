package help

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func New(cmds *[]commands.DiscordCommand) *Command {
	cmd := &Command{
		commands: cmds,
	}

	//nolint:exhaustive // no need to implement everything, only that matters
	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: cmd.getHelp,
		discordgo.InteractionMessageComponent:   cmd.updateHelp,
	}
	return cmd
}

func (command *Command) GetName() string {
	return contract.HelpCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			CommandID:   "</help:1190612462194663555>",
			Description: i18n.Get(lg, "help.help.detailed"),
			TutorialURL: i18n.Get(lg, "help.help.tutorial"),
		},
	}
}

func (command *Command) Matches(i *discordgo.InteractionCreate) bool {
	return command.matchesApplicationCommand(i) || matchesMessageCommand(i)
}

func (command *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command.CallHandler(s, i, command.handlers)
}

func (command *Command) getHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_, err := s.InteractionResponseEdit(i.Interaction, command.getHelpMenu(i.Locale))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle help reponse")
	}
}

func (command *Command) updateHelp(s *discordgo.Session,
	i *discordgo.InteractionCreate) {
	values := i.MessageComponentData().Values
	if len(values) != 1 {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, i.MessageComponentData().CustomID).
			Msgf("Cannot retrieve command name from value, panicking...")
		panic(commands.ErrInvalidInteraction)
	}
	commandName := values[0]

	var webhookEdit *discordgo.WebhookEdit
	if commandName == menuCommandName {
		webhookEdit = command.getHelpMenu(i.Locale)
	} else {
		webhookEdit = command.getHelpCommand(commandName, i.Locale)
	}

	_, err := s.InteractionResponseEdit(i.Interaction, webhookEdit)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle help reponse")
	}
}

func (command *Command) getHelpMenu(lg discordgo.Locale) *discordgo.WebhookEdit {
	type i18nCommand struct {
		Name        string
		Description string
	}

	i18nCommands := make([]i18nCommand, 0)
	for _, command := range *command.commands {
		commandName := command.GetName()
		i18nCommands = append(i18nCommands, i18nCommand{
			Name:        commandName,
			Description: i18n.Get(lg, commandName+".help.overview"),
		})
	}

	sort.SliceStable(i18nCommands, func(i, j int) bool {
		return i18nCommands[i].Name < i18nCommands[j].Name
	})

	return &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       i18n.Get(lg, "help.commands.title"),
				Description: i18n.Get(lg, "help.commands.description", i18n.Vars{"commands": i18nCommands}),
				Color:       constants.Color,
				Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: constants.AvatarIcon},
				Footer:      discord.BuildDefaultFooter(lg),
			},
		},
		Components: command.getHelpComponents(menuCommandName, lg),
	}
}

func (command *Command) getHelpCommand(commandName string, lg discordgo.Locale) *discordgo.WebhookEdit {
	var wantedCommand commands.DiscordCommand
	for _, cmd := range *command.commands {
		if cmd.GetName() == commandName {
			wantedCommand = cmd
			break
		}
	}

	if wantedCommand == nil {
		panic(fmt.Errorf("cannot find command (%s) while searching for its help description, panicking", commandName))
	}

	return &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title: i18n.Get(lg, "help.command.title", i18n.Vars{"command": commandName}),
				Description: i18n.Get(lg, "help.command.description", i18n.Vars{
					"descriptions": wantedCommand.GetDescriptions(lg),
					"source":       i18n.Get(lg, commandName+".help.source"),
				}),
				Color:     constants.Color,
				Image:     &discordgo.MessageEmbedImage{URL: i18n.Get(lg, commandName+".help.tutorial")},
				Thumbnail: &discordgo.MessageEmbedThumbnail{URL: constants.AvatarIcon},
				Footer:    discord.BuildDefaultFooter(lg),
			},
		},
		Components: command.getHelpComponents(commandName, lg),
	}
}

func (command *Command) getHelpComponents(selectedCommandName string, lg discordgo.Locale,
) *[]discordgo.MessageComponent {
	commandChoices := make([]discordgo.SelectMenuOption, 0)
	commandChoices = append(commandChoices, discordgo.SelectMenuOption{
		Label:   i18n.Get(lg, "help.commands.choices.menu"),
		Value:   menuCommandName,
		Default: selectedCommandName == menuCommandName,
		Emoji: &discordgo.ComponentEmoji{
			Name: "📜",
		},
	})

	for _, command := range *command.commands {
		commandName := command.GetName()
		commandChoices = append(commandChoices, discordgo.SelectMenuOption{
			Label:   i18n.Get(lg, "help.commands.choices.command", i18n.Vars{"command": commandName}),
			Value:   commandName,
			Default: selectedCommandName == commandName,
		})
	}

	return &[]discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    contract.CraftHelpCustomID(),
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: i18n.Get(lg, "help.commands.choices.placeholder"),
					Options:     commandChoices,
				},
			},
		},
	}
}

func (command *Command) matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToHelp(i.MessageComponentData().CustomID)
}
