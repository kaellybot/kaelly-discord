package constants

import amqp "github.com/kaellybot/kaelly-amqp"

type AnkamaGame struct {
	Name     string
	Icon     string
	Emoji    string
	AMQPGame amqp.Game
}

const (
	AnkamaLogo = "https://i.imgur.com/dcqcAr2.png"
)

func GetGame() AnkamaGame {
	return GetDofusGame()
}

func GetDofusGame() AnkamaGame {
	return AnkamaGame{
		Name:     "DOFUS",
		Icon:     "https://i.imgur.com/duP1rhM.png",
		Emoji:    "<:dofus:1291317932961304606>", // from KaellyBot server
		AMQPGame: amqp.Game_DOFUS_GAME,
	}
}

func GetDofusTouchGame() AnkamaGame {
	return AnkamaGame{
		Name:     "DOFUS Touch",
		Icon:     "https://i.imgur.com/lYLm648.png",
		Emoji:    "<:dofustouch:1065724958203981944>", // from KaellyBot server
		AMQPGame: amqp.Game_DOFUS_TOUCH,
	}
}

func GetDofusRetroGame() AnkamaGame {
	return AnkamaGame{
		Name:     "DOFUS Retro",
		Icon:     "https://i.imgur.com/PagFd6V.png",
		Emoji:    "<:dofusretro:1065724870392041483>", // from KaellyBot server
		AMQPGame: amqp.Game_DOFUS_RETRO,
	}
}
