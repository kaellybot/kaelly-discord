package constants

type AnkamaGame struct {
	Name  string
	Icon  string
	Emoji string
}

var (
	AnkamaGameDofus = AnkamaGame{
		Name:  "DOFUS",
		Icon:  "https://static.ankama.com/dofus/ng/img/logo_dofus.jpg",
		Emoji: "<:dofus:1065724887525773353>",
	}

	AnkamaGameDofusTouch = AnkamaGame{
		Name:  "DOFUS Touch",
		Icon:  "https://static.ankama.com/dofus-touch/www/img/logo.jpg",
		Emoji: "<:dofustouch:1065724958203981944>",
	}

	AnkamaGameDofusRetro = AnkamaGame{
		Name:  "DOFUS Retro",
		Icon:  "https://static.ankama.com/dofus/ng/modules/mmorpg/retro/new/assets/logo.png",
		Emoji: "<:dofusretro:1065724870392041483>",
	}
)
