package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/rs/zerolog/log"
)

func (service *Impl) welcomeGuild(s *discordgo.Session, guild *discordgo.Guild) {
	// Check if the system channel exists and has the required permissions
	if guild.SystemChannelID != "" &&
		discord.HasPermissions(s, guild.SystemChannelID, welcomeMessagePermissions) {
		sendWelcomeMessage(s, guild, guild.SystemChannelID)
		return
	}

	// Check other channels in the guild
	for _, channel := range guild.Channels {
		if channel.Type == discordgo.ChannelTypeGuildText &&
			discord.HasPermissions(s, channel.ID, welcomeMessagePermissions) {
			sendWelcomeMessage(s, guild, channel.ID)
			return
		}
	}

	// If no suitable channel is found, try to DM the guild owner
	if guild.OwnerID != "" {
		sendWelcomeMessageToOwner(s, guild)
	}
}

func sendWelcomeMessage(s *discordgo.Session, guild *discordgo.Guild, channelID string) {
	message := mappers.MapWelcome(guild.Name, discordgo.Locale(guild.PreferredLocale))
	_, errSend := s.ChannelMessageSendEmbed(channelID, message)
	if errSend != nil {
		log.Warn().Err(errSend).
			Str(constants.LogGuildID, guild.ID).
			Msgf("Failed to send welcome message to channel, ignoring...")
	}
}

func sendWelcomeMessageToOwner(s *discordgo.Session, guild *discordgo.Guild) {
	channel, errCreate := s.UserChannelCreate(guild.OwnerID)
	if errCreate != nil {
		log.Warn().Err(errCreate).
			Str(constants.LogGuildID, guild.ID).
			Msgf("Failed to create DM channel with owner, ignoring...")
		return
	}

	lg := constants.DefaultLocale
	if user, errUser := s.User(guild.OwnerID); errUser == nil {
		lg = discordgo.Locale(user.Locale)
	}

	message := mappers.MapWelcome(guild.Name, lg)
	_, errSend := s.ChannelMessageSendEmbed(channel.ID, message)
	if errSend != nil {
		log.Warn().Err(errSend).
			Str(constants.LogGuildID, guild.ID).
			Msgf("Failed to send welcome message to owner, ignoring...")
	}
}
