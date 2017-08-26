package models

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	TGID      int        `gorm:"primary_key" json:"-"`
	DeletedAt *time.Time `json:"-"`
	CreatedAt time.Time  `json:"-"`
	Username  string     `gorm:"size:64" json:"username"`
	Chats     []Chat     `gorm:"many2many:users_chats;AssociationForeignKey:TGID;ForeignKey:TGID"`
	Token     []Token    `gorm:"ForeignKey:ID;AssociationForeignKey:TGID"`
}

func (u *User) GetTags(db *gorm.DB) ([]Tag, error) {
	tags := []Tag{}

	db.Model(&User{}).
		Preload("Chats.Tags").
		Where("tg_id = ?", u.TGID).
		Find(u)

	for _, tg := range u.Chats {
		tags = append(tags, tg.Tags...)
	}
	return tags, db.Error
}

func (u *User) Register(db *gorm.DB) (int64, error) {
	if err := db.Create(u).Error; err != nil {
		return db.RowsAffected, err
	} else {
		return db.RowsAffected, nil
	}
}

func (u *User) CreateIfNotExists(db *gorm.DB) error {
	err := db.Set("gorm:insert_option", "ON CONFLICT ON CONSTRAINT users_pkey DO NOTHING").
		Create(u).
		Error
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

func (u *User) IsExists(db *gorm.DB) bool {
	ok := db.Where(u).First(u).RowsAffected
	return ok == 1
}

func (u *User) GetUsersChats(db *gorm.DB) error {
	db.Model(&User{}).
		Preload("Chats").
		Where("tg_id = ?", u.TGID).
		Find(u)
	if db.Error != nil {
		return db.Error
	}
	return nil
}
