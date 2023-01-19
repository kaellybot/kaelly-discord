package constants

type AnkamaGame struct {
	Name  string
	Icon  string
	Emoji string
}

const (
	AnkamaLogo = "https://i.imgur.com/dcqcAr2.png"
)

var (
	AnkamaGameDofus = AnkamaGame{
		Name:  "DOFUS",
		Icon:  "https://i.imgur.com/n3fJCSu.png",
		Emoji: "<:dofus:1065724887525773353>",
	}

	AnkamaGameDofusTouch = AnkamaGame{
		Name:  "DOFUS Touch",
		Icon:  "https://i.imgur.com/lYLm648.png",
		Emoji: "<:dofustouch:1065724958203981944>",
	}

	AnkamaGameDofusRetro = AnkamaGame{
		Name:  "DOFUS Retro",
		Icon:  "https://i.imgur.com/PagFd6V.png",
		Emoji: "<:dofusretro:1065724870392041483>",
	}
)
