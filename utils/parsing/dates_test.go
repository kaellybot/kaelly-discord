package parsing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDate(t *testing.T) {
	testCases := []struct {
		name         string
		dateInput    string
		expectedDate *time.Time
		expectedErr  error
	}{
		{
			name:         "Valid little endian date format (1)",
			dateInput:    "20-02-1994",
			expectedDate: createDate(1994, time.February, 20),
			expectedErr:  nil,
		},
		{
			name:         "Valid little endian date format (2)",
			dateInput:    "28/02/1996",
			expectedDate: createDate(1996, time.February, 28),
			expectedErr:  nil,
		},
		{
			name:         "Valid little endian lazy date format",
			dateInput:    "01 09 2004",
			expectedDate: createDate(2004, time.September, 1),
			expectedErr:  nil,
		},
		{
			name:         "Valid big endian date format (1)",
			dateInput:    "1994-02-20",
			expectedDate: createDate(1994, time.February, 20),
			expectedErr:  nil,
		},
		{
			name:         "Valid big endian date format (2)",
			dateInput:    "1996/02/28",
			expectedDate: createDate(1996, time.February, 28),
			expectedErr:  nil,
		},
		{
			name:         "Valid big endian lazy date format",
			dateInput:    "2004 09 01",
			expectedDate: createDate(2004, time.September, 1),
			expectedErr:  nil,
		},
		{
			name:         "Empty date format",
			dateInput:    "",
			expectedDate: nil,
			expectedErr:  ErrDateParsing,
		},
		{
			name:         "Not valid (very) lazy date format",
			dateInput:    "01072010",
			expectedDate: nil,
			expectedErr:  ErrDateParsing,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDate(tc.dateInput)

			assert.Equal(t, tc.expectedDate, result)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

// helper function to create time.Time objects easily.
func createDate(year int, month time.Month, day int) *time.Time {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &date
}
