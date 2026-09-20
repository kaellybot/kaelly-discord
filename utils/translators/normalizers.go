package translators

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NormalizeName lowercases name and strips its diacritics, so that searches
// match whatever the case and accents the user typed.
//
// The transformer is built per call on purpose: transform.Chain keeps internal
// state and resets it on every use, so sharing one across the goroutines
// discordgo spawns per event corrupts the normalized text.
func NormalizeName(name string) (string, error) {
	transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	cleanedName, _, err := transform.String(transformer, strings.ToLower(name))
	return cleanedName, err
}
