package translators

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeName_LowercasesAndStripsDiacritics(t *testing.T) {
	testCases := map[string]string{
		"Draconiros": "draconiros",
		"HÉNUAL":     "henual",
		"Crâpaud":    "crapaud",
		"Cœur":       "coeur", // ligatures are expanded so "coeur" matches it
		"CŒUR":       "coeur",
		"Æon-Nævius": "aeon-naevius",
		"":           "",
	}
	for input, expected := range testCases {
		cleaned, err := NormalizeName(input)
		require.NoError(t, err)
		assert.Equal(t, expected, cleaned)
	}
}

func TestNormalizeName_IsSafeFromConcurrentGoroutines(t *testing.T) {
	// discordgo runs every event handler in its own goroutine, so autocompletes
	// normalize names concurrently. A transformer shared between them keeps
	// internal state and silently corrupts the result.
	const goroutines = 32
	var waitGroup sync.WaitGroup
	results := make([]string, goroutines)

	for index := range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			cleaned, err := NormalizeName("Hénual-Crâpaud Ìmagiro")
			assert.NoError(t, err)
			results[index] = cleaned
		}()
	}
	waitGroup.Wait()

	for _, got := range results {
		assert.Equal(t, "henual-crapaud imagiro", got)
	}
}
