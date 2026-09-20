package constants

import (
	"slices"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSeason_BoundariesAndYearWrap(t *testing.T) {
	cases := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{"new year is winter", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "Winter"},
		{"leap day is winter", time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC), "Winter"},
		{"midnight of March 20 is still winter (start is exclusive)",
			time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC), "Winter"},
		{"March 20 during the day is spring (spring is checked before winter)",
			time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC), "Spring"},
		{"mid spring", time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC), "Spring"},
		{"June 20 during the day is spring", time.Date(2023, 6, 20, 18, 0, 0, 0, time.UTC), "Spring"},
		{"midnight of June 21 is summer", time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC), "Summer"},
		{"mid summer", time.Date(2023, 8, 15, 0, 0, 0, 0, time.UTC), "Summer"},
		{"midnight of September 21 is autumn", time.Date(2023, 9, 21, 0, 0, 0, 0, time.UTC), "Autumn"},
		{"December 20 during the day is autumn", time.Date(2023, 12, 20, 10, 0, 0, 0, time.UTC), "Autumn"},
		{"midnight of December 21 is winter", time.Date(2023, 12, 21, 0, 0, 0, 0, time.UTC), "Winter"},
		{"new year's eve is winter", time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC), "Winter"},
		{"non-UTC input is converted first",
			time.Date(2023, 6, 21, 1, 0, 0, 0, time.FixedZone("UTC+2", 2*60*60)), "Spring"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, GetSeason(tc.date).Name)
		})
	}
}

func TestGetSeason_EveryHourOfAYearHasAKnownSeason(t *testing.T) {
	unknown := GetUnknownSeason().Name
	for date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC); date.Year() == 2024; date = date.Add(time.Hour) {
		if season := GetSeason(date); season.Name == unknown {
			t.Fatalf("no season for %v", date)
		}
	}
}

func TestGetSeason_CarriesSeasonDisplayAttributes(t *testing.T) {
	summer := GetSeason(time.Date(2023, 7, 14, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, summerColor, summer.Color)
	assert.Contains(t, summer.AlmanaxIcon, "summer")

	unknown := GetUnknownSeason()
	assert.Equal(t, "Unknown", unknown.Name)
	assert.Equal(t, Color, unknown.Color)
	assert.Contains(t, unknown.AlmanaxIcon, "unknown")
}

func TestSupportedDateFormats_IncludeDiscordFormatAndRoundTrip(t *testing.T) {
	formats := SupportedDateFormats()
	assert.Contains(t, formats, DiscordDateOnlyFormat, "what the bot prints must be parsable back")

	date := time.Date(2012, 9, 18, 0, 0, 0, 0, time.UTC)
	for _, format := range formats {
		parsed, err := time.Parse(format, date.Format(format))
		require.NoError(t, err, format)
		assert.True(t, date.Equal(parsed), format)
	}
}

func TestAlmanaxBounds(t *testing.T) {
	assert.True(t, GetAlmanaxFirstDate().Before(GetAlmanaxLastDate()))
	assert.Equal(t, "2012-09-18", GetAlmanaxFirstDate().Format(KrosmozAlmanaxDateFormat))
	assert.Equal(t, time.UTC, GetAlmanaxFirstDate().Location())
}

func TestAutocompleteChoices_FitDiscordLimitAndAreSortedUnique(t *testing.T) {
	for name, values := range map[string][]int64{
		"almanax day durations": GetAlmanaxDayDuration(),
		"character numbers":     GetCharacterNumbers(),
	} {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, values)
			assert.LessOrEqual(t, len(values), MaxChoices, "Discord refuses more than 25 choices")
			assert.True(t, slices.IsSorted(values))
			assert.Len(t, slices.Compact(slices.Clone(values)), len(values), "no duplicates")
			assert.Positive(t, values[0])
		})
	}
	durations := GetAlmanaxDayDuration()
	assert.Equal(t, int64(7), durations[0], "a week at least")
	assert.Equal(t, int64(30), durations[len(durations)-1])
}

func TestGetRabbitMQClientID_SuffixedWithShardID(t *testing.T) {
	previous := viper.Get(DiscordShardID)
	t.Cleanup(func() { viper.Set(DiscordShardID, previous) })

	viper.Set(DiscordShardID, 3)
	assert.Equal(t, "KaellyBot-3", GetRabbitMQClientID())

	viper.Set(DiscordShardID, nil)
	assert.Equal(t, "KaellyBot-", GetRabbitMQClientID())
}

func TestGetDefaultConfigValues_CoversEveryConfigKeyWithUsableDefaults(t *testing.T) {
	defaults := GetDefaultConfigValues()

	for _, key := range []string{DiscordToken, DiscordShardID, DiscordShardCount, MySQLURL, MySQLUser,
		MySQLPassword, MySQLDatabase, RabbitMQAddress, AboutID, AlignID, AlmanaxID, ConfigID, HelpID,
		ItemID, JobID, MapID, SetID, ProbePort, MetricPort, LogLevel, Production,
		RequestTimeout} {
		assert.Contains(t, defaults, key)
	}
	assert.Len(t, defaults, 22)

	assert.Equal(t, 1, defaults[DiscordShardCount], "a single shard by default")
	assert.Equal(t, false, defaults[Production], "never register commands globally by accident")
	assert.NotEqual(t, defaults[ProbePort], defaults[MetricPort], "both servers bind at once")
	assert.Equal(t, 60, defaults[RequestTimeout], "well below Discord's 15-minute token lifetime")

	rawLevel, ok := defaults[LogLevel].(string)
	require.True(t, ok)
	level, err := zerolog.ParseLevel(rawLevel)
	require.NoError(t, err)
	assert.Equal(t, zerolog.InfoLevel, level)
}

func TestGetIntents_NoPrivilegedMessageContent(t *testing.T) {
	intents := GetIntents()

	assert.Zero(t, intents&discordgo.IntentMessageContent, "message content intent was removed on purpose")
	for _, required := range []discordgo.Intent{discordgo.IntentGuildMembers, discordgo.IntentGuilds,
		discordgo.IntentGuildMessages, discordgo.IntentGuildMessageReactions, discordgo.IntentGuildWebhooks} {
		assert.Equal(t, required, intents&required)
	}
}

func TestGetGame_IsDofus(t *testing.T) {
	game := GetGame()
	assert.Equal(t, "DOFUS", game.Name)
	assert.NotEmpty(t, game.Icon)
	assert.EqualValues(t, "DOFUS_GAME", game.AMQPGame.String())
}

func TestGraphists_HaveNameAndLink(t *testing.T) {
	for _, graphist := range []Graphist{GetGraphistElycann(), GetGraphistColibry()} {
		assert.NotEmpty(t, graphist.Name)
		assert.Regexp(t, `^https://`, graphist.URL)
	}
}

// A missing or non-positive timeout would expire every request the instant it is
// made, leaving every command answering "try again later".
func TestGetRequestTimeout_FallsBackOnUnusableValues(t *testing.T) {
	previous := viper.Get(RequestTimeout)
	t.Cleanup(func() { viper.Set(RequestTimeout, previous) })

	for _, value := range []any{nil, 0, -1} {
		viper.Set(RequestTimeout, value)
		assert.Equal(t, 60*time.Second, GetRequestTimeout())
	}

	viper.Set(RequestTimeout, 5)
	assert.Equal(t, 5*time.Second, GetRequestTimeout())
}
