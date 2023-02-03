package constants

import "github.com/kaellybot/kaelly-discord/models/entities"

const (
	AlignmentMinLevel = 0
	AlignmentMaxLevel = 100
)

var (
	NeutralCity = entities.City{
		Color: 12506502,
		Icon:  "https://i.imgur.com/i74Rh8o.png",
	}
)

type AlignmentUserLevel struct {
	CityId   string
	OrderId  string
	Username string
	Level    int64
}
