package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	//gorm.Model
	ID       uint          `gorm:"primary_key;AUTO_INCREMENT"`
	Username string        `gorm:"size:64"`
	Password string        `gorm:"type:varchar(128)"`
	Email    string        `gorm:"unique"`
	//Chat []Chat				`gorm:"many2many:users_chats;"`
}

func (u *User) Register(db *gorm.DB) (int64, error) {
	if err := db.Create(u).Error; err != nil {
		return db.RowsAffected, err
	} else {
		return db.RowsAffected, nil
	}
}
