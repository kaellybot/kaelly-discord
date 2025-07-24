package constants

import "time"

const (
	KrosmozAlmanaxDateFormat = "2006-01-02"
)

func GetAlmanaxFirstDate() time.Time {
	return time.Date(2012, 9, 18, 0, 0, 0, 0, time.UTC)
}

func GetAlmanaxLastDate() time.Time {
	return time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
}

// 25 values with most used days duration.
func GetAlmanaxDayDuration() []int64 {
	return []int64{
		7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
	}
}

// 25 values with most used character numbers.
func GetCharacterNumbers() []int64 {
	return []int64{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		25, 30, 40, 50, 100,
	}
}
