package models

import (
	"time"
)

type Image struct {
	ID uint64 `gorm:"primary_key;AUTO_INCREMENT"`
	//Src    string
	ChatID uint64
	Tags   []Tag `sql:"many2many:images_tags;"`
	Date   time.Time
}
