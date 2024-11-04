package constants

import (
	"time"
)

const (
	springColor = 9358615
	summerColor = 16776960
	autumnColor = 15301376
	winterColor = 50912
)

type Season struct {
	Name        string
	StartDate   time.Time // exclusive
	EndDate     time.Time // exclusive
	Color       int
	AlmanaxIcon string
}

func GetSeason(date time.Time) Season {
	utcDate := date.UTC()
	for _, season := range getSeasons() {
		seasonPreviousStartDate := season.StartDate.AddDate(utcDate.Year()-1, 0, 0)
		seasonPreviousEndDate := season.EndDate.AddDate(utcDate.Year()-1, 0, 0)
		seasonStartDate := season.StartDate.AddDate(utcDate.Year(), 0, 0)
		seasonEndDate := season.EndDate.AddDate(utcDate.Year(), 0, 0)
		if seasonPreviousStartDate.Before(utcDate) && seasonPreviousEndDate.After(utcDate) ||
			seasonStartDate.Before(utcDate) && seasonEndDate.After(utcDate) {
			return season
		}
	}

	return GetUnknownSeason()
}

func GetUnknownSeason() Season {
	return Season{
		Name:        "Unknown",
		Color:       Color,
		AlmanaxIcon: "https://raw.githubusercontent.com/KaellyBot/Kaelly-cdn/refs/heads/main/common/almanax/unknown.webp",
	}
}

func getSeasons() []Season {
	return []Season{
		{
			Name:        "Spring",
			StartDate:   time.Date(0, 3, 20, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(0, 6, 21, 0, 0, 0, 0, time.UTC),
			Color:       springColor,
			AlmanaxIcon: "https://raw.githubusercontent.com/KaellyBot/Kaelly-cdn/refs/heads/main/common/almanax/spring.webp",
		},
		{
			Name:        "Summer",
			StartDate:   time.Date(0, 6, 20, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(0, 9, 21, 0, 0, 0, 0, time.UTC),
			Color:       summerColor,
			AlmanaxIcon: "https://raw.githubusercontent.com/KaellyBot/Kaelly-cdn/refs/heads/main/common/almanax/summer.webp",
		},
		{
			Name:        "Autumn",
			StartDate:   time.Date(0, 9, 20, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(0, 12, 21, 0, 0, 0, 0, time.UTC),
			Color:       autumnColor,
			AlmanaxIcon: "https://raw.githubusercontent.com/KaellyBot/Kaelly-cdn/refs/heads/main/common/almanax/autumn.webp",
		},
		{
			Name:        "Winter",
			StartDate:   time.Date(0, 12, 20, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(1, 3, 21, 0, 0, 0, 0, time.UTC),
			Color:       winterColor,
			AlmanaxIcon: "https://raw.githubusercontent.com/KaellyBot/Kaelly-cdn/refs/heads/main/common/almanax/winter.webp",
		},
	}
}
