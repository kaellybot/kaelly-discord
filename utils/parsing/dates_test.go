package parsing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestParseDate_RejectsImpossibleOrAmbiguousInputs(t *testing.T) {
	for _, input := range []string{
		"31/02/2024", // no such day
		"29/02/2023", // not a leap year
		"20-02/1994", // mixed separators
		"1994-20-02", // big endian with month/day swapped
		"20/02/94",   // two-digit year
		" 20/02/1994",
		"20/02/1994 12:00",
	} {
		t.Run(input, func(t *testing.T) {
			date, err := ParseDate(input)
			require.ErrorIs(t, err, ErrDateParsing)
			assert.Nil(t, date)
		})
	}
}

func TestParseDate_LeapDayIsMidnightUTC(t *testing.T) {
	date, err := ParseDate("29/02/2024")

	require.NoError(t, err)
	assert.Equal(t, time.UTC, date.Location())
	assert.Equal(t, time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC), *date)
}

// helper function to create time.Time objects easily.
func createDate(year int, month time.Month, day int) *time.Time {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &date
}
