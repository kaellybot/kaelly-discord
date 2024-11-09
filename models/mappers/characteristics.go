package mappers

import (
	"fmt"
	"math"
	"regexp"
	"sort"

	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
)

type AMQPCharacteristic interface {
	GetId() string
	GetLabel() string
}

type i18nCharacteristic struct {
	Label     string
	Emoji     string
	SortOrder int
}

var (
	effectValue = regexp.MustCompile(`(-?\d+)`)
)

func mapEffects[Characteristic AMQPCharacteristic](effects []Characteristic,
	characteristicService characteristics.Service) []i18nCharacteristic {
	characs := make([]i18nCharacteristic, 0)
	for _, effect := range effects {
		charac, found := characteristicService.GetCharacteristic(effect.GetId())
		if !found {
			charac = entities.Characteristic{
				SortOrder: math.MaxInt,
			}
		}

		characs = append(characs, i18nCharacteristic{
			Label:     highlightEffectValues(effect.GetLabel()),
			Emoji:     charac.Emoji,
			SortOrder: charac.SortOrder,
		})
	}

	sortCharacteristics(characs)
	return characs
}

func highlightEffectValues(label string) string {
	return effectValue.ReplaceAllStringFunc(label, func(match string) string {
		return fmt.Sprintf("**%s**", match)
	})
}

func sortCharacteristics(characteristics []i18nCharacteristic) {
	sort.SliceStable(characteristics, func(i, j int) bool {
		return characteristics[i].SortOrder < characteristics[j].SortOrder
	})
}
