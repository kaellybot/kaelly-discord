package discord

import "github.com/bwmarrin/discordgo"

type ItemsToField[T any] func(i int, items []T) *discordgo.MessageEmbedField
type ItemToButton[T any] func(item T) discordgo.Button
