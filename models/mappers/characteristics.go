package mappers

import (
	"math"
	"sort"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
)

type i18nCharacteristic struct {
	Label     string
	Emoji     string
	SortOrder int
}

func mapSetEffects(effects []*amqp.EncyclopediaSetAnswer_Effect,
	characteristicService characteristics.Service) []i18nCharacteristic {
	characs := make([]i18nCharacteristic, 0)
	for _, effect := range effects {
		charac, found := characteristicService.GetCharacteristic(effect.Id)
		if !found {
			charac = entities.Characteristic{
				SortOrder: math.MaxInt,
			}
		}

		characs = append(characs, i18nCharacteristic{
			Label:     effect.Label,
			Emoji:     charac.Emoji,
			SortOrder: charac.SortOrder,
		})
	}

	sortCharacteristics(characs)
	return characs
}

func sortCharacteristics(characteristics []i18nCharacteristic) {
	sort.SliceStable(characteristics, func(i, j int) bool {
		return characteristics[i].SortOrder < characteristics[j].SortOrder
	})
}
