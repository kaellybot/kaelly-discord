package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/mappers"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/rs/zerolog/log"
)

func (service *Impl) welcomeGuild(guild *discordgo.Guild) {
	// Check if the system channel exists and has the required permissions
	if guild.SystemChannelID != "" &&
		discord.HasPermissions(service.session, guild.SystemChannelID, welcomeMessagePermissions) &&
		service.sendWelcomeMessage(guild, guild.SystemChannelID) {
		return
	}

	// Check other channels in the guild
	for _, channel := range guild.Channels {
		if channel.ID != guild.SystemChannelID &&
			channel.Type == discordgo.ChannelTypeGuildText &&
			discord.HasPermissions(service.session, channel.ID, welcomeMessagePermissions) &&
			service.sendWelcomeMessage(guild, channel.ID) {
			return
		}
	}

	// If no channel accepted the message, try to DM the guild owner
	if guild.OwnerID != "" {
		service.sendWelcomeMessageToOwner(guild)
	}
}

func (service *Impl) sendWelcomeMessage(guild *discordgo.Guild, channelID string) bool {
	lg := discordgo.Locale(guild.PreferredLocale)
	message := mappers.MapWelcomeToEmbed(guild.Name, guild.OwnerID, lg, service.emojiService)
	_, errSend := service.session.ChannelMessageSendEmbed(channelID, message)
	if errSend != nil {
		log.Warn().Err(errSend).
			Str(constants.LogGuildID, guild.ID).
			Msgf("Failed to send welcome message to channel, trying another target...")
		return false
	}

	return true
}

func (service *Impl) sendWelcomeMessageToOwner(guild *discordgo.Guild) {
	channel, errCreate := service.session.UserChannelCreate(guild.OwnerID)
	if errCreate != nil {
		log.Warn().Err(errCreate).
			Str(constants.LogGuildID, guild.ID).
			Msgf("Failed to create DM channel with owner, ignoring...")
		return
	}

	// Discord does not return the locale of a user fetched by a bot,
	// so the guild preferred locale is the best fallback available.
	lg := discordgo.Locale(guild.PreferredLocale)
	if user, errUser := service.session.User(guild.OwnerID); errUser == nil && user.Locale != "" {
		lg = discordgo.Locale(user.Locale)
	}

	message := mappers.MapWelcomeToEmbed(guild.Name, guild.OwnerID, lg, service.emojiService)
	_, errSend := service.session.ChannelMessageSendEmbed(channel.ID, message)
	if errSend != nil {
		log.Warn().Err(errSend).
			Str(constants.LogGuildID, guild.ID).
			Msgf("Failed to send welcome message to owner, ignoring...")
	}
}
