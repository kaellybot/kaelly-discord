package help

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/middlewares"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/rs/zerolog/log"
)

func New(broker amqp.MessageBroker, cmds *[]commands.DiscordCommand) *Command {
	cmd := &Command{
		broker:   broker,
		commands: cmds,
	}

	//nolint:exhaustive // no need to implement everything, only that matters
	cmd.handlers = commands.DiscordHandlers{
		discordgo.InteractionApplicationCommand: middlewares.Use(cmd.trace, cmd.getHelp),
		discordgo.InteractionMessageComponent:   middlewares.Use(cmd.trace, cmd.updateHelp),
	}
	return cmd
}

func (command *Command) GetName() string {
	return contract.HelpCommandName
}

func (command *Command) GetDescriptions(lg discordgo.Locale) []commands.Description {
	return []commands.Description{
		{
			Name:        "/help",
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

func (command *Command) trace(ctx context.Context, _ *discordgo.Session,
	i *discordgo.InteractionCreate, next middlewares.NextFunc) {
	authorID := discord.GetUserID(i.Interaction)
	message := mappers.MapHelpRequest(authorID, i.Locale)
	errBroker := command.broker.Request(message, amqp.ExchangeRequest, routingKey, i.ID, requests.AnswersQueueName)
	if errBroker != nil {
		log.Error().Err(errBroker).Msgf("Cannot trace help interaction through AMQP")
	}

	next(ctx)
}

func (command *Command) getHelp(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	_, err := s.InteractionResponseEdit(i.Interaction, command.getHelpMenu(i.Locale))
	if err != nil {
		log.Error().Err(err).Msgf("Cannot handle help reponse")
	}
}

func (command *Command) updateHelp(_ context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, _ middlewares.NextFunc) {
	var commandName string
	var page int
	customID := i.MessageComponentData().CustomID
	values := i.MessageComponentData().Values
	if len(values) != 1 {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, i.MessageComponentData().CustomID).
			Msgf("Cannot retrieve command name from value, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	if customCmdName, ok := contract.ExtractHelpPageCustomID(customID); ok {
		commandName = customCmdName
		customPage, err := strconv.Atoi(values[0])
		if err != nil {
			log.Error().
				Err(err).
				Str(constants.LogCommand, command.GetName()).
				Str(constants.LogCustomID, customID).
				Msgf("Cannot convert page value to int, panicking...")
			panic(commands.ErrInvalidInteraction)
		}
		page = customPage
	} else if contract.IsBelongsToHelp(customID) {
		commandName = values[0]
		page = 0
	} else {
		log.Error().
			Str(constants.LogCommand, command.GetName()).
			Str(constants.LogCustomID, customID).
			Msgf("Cannot handle custom ID, panicking...")
		panic(commands.ErrInvalidInteraction)
	}

	var webhookEdit *discordgo.WebhookEdit
	if commandName == menuCommandName {
		webhookEdit = command.getHelpMenu(i.Locale)
	} else {
		webhookEdit = command.getHelpCommand(commandName, page, i.Locale)
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
			Description: i18n.Get(lg, fmt.Sprintf("%v.help.overview", commandName)),
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
		Components: command.getHelpComponents(menuCommandName, 0, nil, lg),
	}
}

func (command *Command) getHelpCommand(commandName string, page int,
	lg discordgo.Locale) *discordgo.WebhookEdit {
	var wantedCommand commands.DiscordCommand
	for _, cmd := range *command.commands {
		if cmd.GetName() == commandName {
			wantedCommand = cmd
			break
		}
	}

	if wantedCommand == nil {
		panic(fmt.Errorf("cannot find command (%s) while searching"+
			" for its help description, panicking", commandName))
	}

	details := wantedCommand.GetDescriptions(lg)
	detail := details[page]
	commandID := detail.CommandID
	var subCommandID string
	if len(details) > 1 {
		commandID = commandName
		subCommandID = detail.CommandID
	}

	return &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title: i18n.Get(lg, "help.command.title", i18n.Vars{
					"command":   commandID,
					"commandID": subCommandID,
				}),
				Description: i18n.Get(lg, "help.command.description", i18n.Vars{
					"detail": detail,
					"source": i18n.Get(lg, fmt.Sprintf("%v.help.source", commandName)),
				}),
				Color:     constants.Color,
				Image:     &discordgo.MessageEmbedImage{URL: detail.TutorialURL},
				Thumbnail: &discordgo.MessageEmbedThumbnail{URL: constants.AvatarIcon},
				Footer:    discord.BuildDefaultFooter(lg),
			},
		},
		Components: command.getHelpComponents(commandName, page, details, lg),
	}
}

func (command *Command) getHelpComponents(selectedCommandName string, page int,
	descriptions []commands.Description, lg discordgo.Locale) *[]discordgo.MessageComponent {
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

	components := []discordgo.MessageComponent{
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

	if len(descriptions) > 1 {
		commandPages := make([]discordgo.SelectMenuOption, 0)
		for i, description := range descriptions {
			commandPages = append(commandPages, discordgo.SelectMenuOption{
				Label:   description.Name,
				Value:   fmt.Sprintf("%v", i),
				Default: i == page,
			})
		}

		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    contract.CraftHelpPageCustomID(selectedCommandName),
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: i18n.Get(lg, "help.commands.pages.placeholder"),
					Options:     commandPages,
				},
			},
		})
	}

	return &components
}

func (command *Command) matchesApplicationCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsApplicationCommand(i) &&
		i.ApplicationCommandData().Name == command.GetName()
}

func matchesMessageCommand(i *discordgo.InteractionCreate) bool {
	return commands.IsMessageCommand(i) &&
		contract.IsBelongsToHelp(i.MessageComponentData().CustomID)
}
