package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"time"
)

type Chat struct {
	TGID      int64      `gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
	Title     string     `json:"title"`
	Users     []User     `json:"-"`
	Images    []Image    `gorm:"ForeignKey:TGID;AssociationForeignKey:ChatID" json:"images,omitempty"`
	Tags      []Tag      `gorm:"many2many:chats_tags;AssociationForeignKey:ID;ForeignKey:TGID"`
}

func (ch *Chat) GetTags(db *gorm.DB) ([]Tag, error) {
	db.Model(&Chat{}).
		Preload("Tags").
		Where("tg_id = ?", ch.TGID).
		Find(ch)
	return ch.Tags, db.Error
}

func (ch *Chat) CreateIfNotExists(db *gorm.DB) error {
	err := db.Set("gorm:insert_option", "ON CONFLICT ON CONSTRAINT chats_pkey DO NOTHING").
		Create(ch).
		Error
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}
