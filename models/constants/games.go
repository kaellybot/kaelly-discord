package constants

import amqp "github.com/kaellybot/kaelly-amqp"

type AnkamaGame struct {
	Name     string
	Icon     string
	AMQPGame amqp.Game
}

const (
	AnkamaLogo = "https://i.imgur.com/dcqcAr2.png"
)

func GetGame() AnkamaGame {
	return AnkamaGame{
		Name:     "DOFUS",
		Icon:     "https://i.imgur.com/duP1rhM.png",
		AMQPGame: amqp.Game_DOFUS_GAME,
	}
}
