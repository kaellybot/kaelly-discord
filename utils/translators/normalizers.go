package translators

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NormalizeName lowercases name, expands its Latin ligatures and strips its
// diacritics, so that searches match whatever the case, ligatures and accents
// the user typed.
//
// The transformer is built per call on purpose: transform.Chain keeps internal
// state and resets it on every use, so sharing one across the goroutines
// discordgo spawns per event corrupts the normalized text.
func NormalizeName(name string) (string, error) {
	transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	cleanedName, _, err := transform.String(transformer, expandLigatures(strings.ToLower(name)))
	return cleanedName, err
}

// expandLigatures replaces the Latin ligatures by their two letters. They are
// single runes NFD does not decompose, so without this "coeur" would never
// match "Cœur".
func expandLigatures(name string) string {
	if !strings.ContainsAny(name, "œŒæÆ") {
		return name
	}

	var builder strings.Builder
	builder.Grow(len(name))
	for _, letter := range name {
		switch letter {
		case 'œ', 'Œ':
			builder.WriteString("oe")
		case 'æ', 'Æ':
			builder.WriteString("ae")
		default:
			builder.WriteRune(letter)
		}
	}

	return builder.String()
}
