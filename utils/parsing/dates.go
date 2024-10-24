package parsing

import (
	"errors"
	"time"

	"github.com/kaellybot/kaelly-discord/models/constants"
)

var (
	ErrDateParsing = errors.New("cannot parse date input")
)

func ParseDate(dateInput string) (*time.Time, error) {
	for _, dateFormat := range constants.SupportedDateFormats() {
		date, err := time.ParseInLocation(dateFormat, dateInput, time.UTC)
		if err == nil {
			return &date, nil
		}
	}

	return nil, ErrDateParsing
}
