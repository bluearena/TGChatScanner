package models

import "github.com/jinzhu/gorm"

type Tag struct {
	gorm.Model
	Name   string  `sql:"not null" json:"name"`
	Image  []Image `sql:"many2many:images_tags" json:"-"`
	ChatID uint
}
