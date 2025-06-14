package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
)

//nolint:lll // Clearer like that.
type WeaponAreaEffect struct {
	ID     string `gorm:"primaryKey"`
	Type   constants.WeaponAreaType
	Order  int
	Labels []WeaponAreaEffectLabel `gorm:"foreignKey:WeaponAreaEffectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type WeaponAreaEffectLabel struct {
	WeaponAreaEffectID string        `gorm:"primaryKey"`
	Locale             amqp.Language `gorm:"primaryKey"`
	Label              string
}

func (areaEffect WeaponAreaEffect) GetID() string {
	return areaEffect.ID
}

func (areaEffect WeaponAreaEffect) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range areaEffect.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
