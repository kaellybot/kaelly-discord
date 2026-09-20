package mappers

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapQueryMismatch_EchoesQueryInCallerLocale(t *testing.T) {
	for _, lg := range []discordgo.Locale{discordgo.French, discordgo.EnglishGB} {
		t.Run(string(lg), func(t *testing.T) {
			edit := MapQueryMismatch("gelano", lg)

			require.NotNil(t, edit.Content)
			expected := make([]string, 0)
			for _, variant := range testutils.Translations(t, lg, "errors.query_mismatch") {
				expected = append(expected, strings.ReplaceAll(variant, "{{ .value }}", "gelano"))
			}
			assert.Contains(t, expected, *edit.Content)
			assert.Nil(t, edit.Embeds)
			assert.Nil(t, edit.Components)
		})
	}
}
