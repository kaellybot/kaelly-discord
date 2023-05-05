package constants

import "github.com/kaellybot/kaelly-discord/models/entities"

const (
	AlignmentMinLevel = 0
	AlignmentMaxLevel = 100

	neutralCityColor = 12506502
)

type AlignmentUserLevel struct {
	CityID   string
	OrderID  string
	Username string
	Level    int64
}

func GetNeutralCity() entities.City {
	return entities.City{
		Color: neutralCityColor,
		Icon:  "https://i.imgur.com/i74Rh8o.png",
	}
}
