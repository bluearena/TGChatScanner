package models

import "github.com/jinzhu/gorm"

type Tag struct {
	gorm.Model
	Name  string
	Image []Image `sql:"many2many:images_tags;"`
}
