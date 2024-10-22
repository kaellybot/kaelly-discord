package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/slicers"
	i18n "github.com/kaysoro/discordgo-i18n"
)

func SliceFields[T any](items []T, limit int, toField ItemsToField[T]) []*discordgo.MessageEmbedField {
	fields := make([]*discordgo.MessageEmbedField, 0)
	slicedItems := slicers.Slice(items, limit)
	for i, items := range slicedItems {
		fields = append(fields, toField(i, items))
	}

	return fields
}

func SliceButtons[T any](items []T, toButton ItemToButton[T]) []discordgo.ActionsRow {
	actionsRow := make([]discordgo.ActionsRow, 0)
	slicedItems := slicers.Slice(items, constants.MaxButtonPerActionRow)
	for _, items := range slicedItems {
		buttons := make([]discordgo.MessageComponent, 0)
		for _, subItem := range items {
			buttons = append(buttons, toButton(subItem))
		}
		actionsRow = append(actionsRow, discordgo.ActionsRow{
			Components: buttons,
		})
	}

	return actionsRow
}

func GhostInlineField() *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:   constants.InvisibleCharacter,
		Inline: false,
	}
}

func BuildDefaultFooter(lg discordgo.Locale) *discordgo.MessageEmbedFooter {
	return &discordgo.MessageEmbedFooter{
		Text: i18n.Get(lg, "default.footer", i18n.Vars{
			"name":      constants.Name,
			"version":   constants.Version,
			"changelog": i18n.Get(lg, "default.changelog"),
		}),
		IconURL: constants.AvatarIcon,
	}
}
