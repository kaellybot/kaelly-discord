package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ExtractAPIError(err error) (*discordgo.APIErrorMessage, bool) {
	// Find JSON portion of the error
	startIndex := strings.Index(err.Error(), "{")
	if startIndex == -1 {
		return nil, false
	}

	// Parse JSON error into discordgo.APIErrorMessage
	var discordError discordgo.APIErrorMessage
	jsonPart := err.Error()[startIndex:]
	if errJSON := discordgo.Unmarshal([]byte(jsonPart), &discordError); errJSON != nil {
		return nil, false
	}

	return &discordError, true
}
