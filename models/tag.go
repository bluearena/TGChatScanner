package models

type Tag struct {
	ID    uint `gorm:"primary_key"`
	Name  string
	Image []Image `sql:"many2many:images_tags;"`
}
