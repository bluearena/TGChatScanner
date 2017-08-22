package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	TGID     uint64  `gorm:"primary_key:true" json:"-"`
	Username string  `gorm:"size:64" json:"username"`
	Chats    []Chat  `gorm:"many2many:users_chats;AssociationForeignKey:TGID;ForeignKey:TGID"`
	Token    []Token `gorm:"ForeignKey:ID;AssociationForeignKey:TGID"`
}

func (u *User) Register(db *gorm.DB) (int64, error) {
	if err := db.Create(u).Error; err != nil {
		return db.RowsAffected, err
	} else {
		return db.RowsAffected, nil
	}
}

func (u *User) IsExists(db *gorm.DB) bool {
	ok := db.Where(u).First(u).RowsAffected
	if ok == 1 {
		return true
	}
	return false
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
