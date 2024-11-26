package mappers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/discord"
)

func MapWelcome(guildName string, lg discordgo.Locale) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: "Salut et merci pour l'invitation ! :tada:\n\nMoi c'est **KaellyBot**, ton nouvel assistant Discord dédié au jeu **DOFUS** !\nMon but ? T'offrir des outils pratiques et funs pour pimenter ton expérience de jeu. 🔥\nVoici quelques-uns de mes super-pouvoirs : \n✨ Consulter l'almanax, les équipements et panoplies du jeu\n✨ Découvrir les positions de portails partagées par la communauté\n✨ Gérer l'annuaire des artisans et alignés de ta guilde\n✨ Tirer aléatoirement des cartes compétitives pour défier tes amis\n\nCurieux de voir tout ce dont je suis capable ? Tape `/help` et explore toutes mes commandes ! 😏\n\n<@162842827183751169>, en tant qu'administrateur de **Xx-best-guild-xX**, tu peux accéder à des fonctionnalités avancées pour configurer mes services.\nAvec `/config`, active :\n📅 L'envoi quotidien de l'almanax\n🐦 Les notifications des tweets et flux RSS du jeu\n🌍 Le serveur de jeu principal de ta guilde\n\nÇa promet d'être épique, non ? Hâte de collaborer avec vous pour rendre ce serveur encore plus fun et utile ! 😄",
		Color:       constants.Color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: constants.AvatarIcon,
		},
		Footer: discord.BuildDefaultFooter(lg),
	}
}
