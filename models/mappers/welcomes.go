package mappers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	contract "github.com/kaellybot/kaelly-commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/emojis"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	i18n "github.com/kaysoro/discordgo-i18n"
	"github.com/spf13/viper"
)

func MapWelcomeToEmbed(guildName, ownerID string, lg discordgo.Locale,
	emojiService emojis.Service) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: i18n.Get(lg, "welcome", i18n.Vars{
			"name":     constants.Name,
			"game":     constants.GetGame(),
			"gameLogo": emojiService.GetMiscStringEmoji(constants.EmojiIDGame),
			"help": fmt.Sprintf("</%v:%v>", contract.HelpCommandName,
				viper.GetString(constants.HelpID)),
			"owner": ownerID,
			"guild": guildName,
			"config": fmt.Sprintf("</%v:%v>", contract.ConfigCommandName,
				viper.GetString(constants.ConfigID)),
			"almanax": emojiService.GetMiscStringEmoji(constants.EmojiIDCalendar),
		}),
		Color: constants.Color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: constants.GetGame().Icon,
		},
		Image: &discordgo.MessageEmbedImage{
			URL: constants.AvatarImage,
		},
		Footer: discord.BuildDefaultFooter(lg),
	}
}
