package entities

import "gorm.io/gorm"

type Server struct {
	gorm.Model
	Id string
}
