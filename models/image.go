package models

import "time"

type Image struct {
	Id  uint	`gorm:"primary_key;AUTO_INCREMENT"`
	Src	string
	ChatID uint64
	Tags []Tag   `sql:"many2many:images_tags;"`
	Date time.Time
}