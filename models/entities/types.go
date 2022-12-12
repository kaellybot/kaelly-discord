package entities

import "github.com/bwmarrin/discordgo"

type LabelledEntity interface {
	GetId() string
	GetLabels() map[discordgo.Locale]string
}
