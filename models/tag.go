package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
)

type Tag struct {
	gorm.Model
	Name   string  `gorm:"not null; unique" json:"name"`
	Images []Image `gorm:"many2many:images_tags" json:"-"`
	Chats  []Chat  `gorm:"many2many:chats_tags"`
}

func (t *Tag) SaveIfUnique(db *gorm.DB) error {
	err := db.Set("gorm:insert_option", "ON CONFLICT ON CONSTRAINT tags_name_key DO NOTHING").
		Save(t).
		Error
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}
