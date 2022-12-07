package entities

type Dimension struct {
	Id             string `gorm:"primaryKey"`
	DofusPortalsId string `gorm:"unique"`
	Icon           string
	Color          int
}
