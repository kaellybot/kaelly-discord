package entities

type Characteristic struct {
	ID        string `gorm:"primaryKey"`
	Emoji     string
	SortOrder int
}
