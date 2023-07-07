package mappers

import (
	"math"
	"sort"

	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/services/characteristics"
)

type AMQPCharacteristic interface {
	GetId() string
	GetLabel() string
	GetIsActive() bool
}

type i18nCharacteristic struct {
	Label     string
	Emoji     string
	SortOrder int
	IsActive  bool
}

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
			Label:     effect.GetLabel(),
			Emoji:     charac.Emoji,
			SortOrder: charac.SortOrder,
			IsActive:  effect.GetIsActive(),
		})
	}

	sortCharacteristics(characs)
	return characs
}

func sortCharacteristics(characteristics []i18nCharacteristic) {
	sort.SliceStable(characteristics, func(i, j int) bool {
		if characteristics[i].IsActive == characteristics[j].IsActive {
			return characteristics[i].SortOrder < characteristics[j].SortOrder
		}

		return characteristics[i].IsActive
	})
}
