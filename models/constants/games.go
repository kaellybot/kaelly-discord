package constants

import amqp "github.com/kaellybot/kaelly-amqp"

type AnkamaGame struct {
	Name     string
	Icon     string
	AMQPGame amqp.Game
}

const (
	AnkamaLogo = "https://raw.githubusercontent.com/KaellyBot/Kaelly-cdn/refs/heads/main/common/logos/ankama.webp"
)

func GetGame() AnkamaGame {
	return AnkamaGame{
		Name:     "DOFUS",
		Icon:     "https://raw.githubusercontent.com/KaellyBot/Kaelly-cdn/refs/heads/main/common/logos/dofus.webp",
		AMQPGame: amqp.Game_DOFUS_GAME,
	}
}
