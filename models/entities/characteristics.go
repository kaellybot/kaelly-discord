package entities

type Characteristic struct {
	ID        string `gorm:"primaryKey"`
	DebugName string
	Emoji     string
	SortOrder int `gorm:"unique"`
}
