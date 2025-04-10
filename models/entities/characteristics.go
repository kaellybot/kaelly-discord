package entities

//nolint:lll // Cannot improve lisibility.
type Characteristic struct {
	ID        string                `gorm:"primaryKey"`
	SortOrder int                   `gorm:"unique"`
	Regexes   []RegexCharacteristic `gorm:"foreignKey:CharacteristicID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type RegexCharacteristic struct {
	CharacteristicID         string `gorm:"primaryKey"`
	Expression               string `gorm:"primaryKey"`
	RelativeCharacteristicID string
}
