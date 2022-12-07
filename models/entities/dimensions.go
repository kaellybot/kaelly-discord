package entities

import "gorm.io/gorm"

type Dimension struct {
	gorm.Model
	Id string
}
