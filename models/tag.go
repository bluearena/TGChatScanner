package models

type Tag struct {
	ID uint			`gorm:"primary_key"`
	Name string
}