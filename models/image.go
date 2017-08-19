package models

import "time"

type Image struct {
	Id  uint	`gorm:"primary_key;AUTO_INCREMENT"`
	Src	string
	Chat *Chat
	Tags []Tag   //`gorm:"many2many:images_tags;"`
	Date time.Time
}